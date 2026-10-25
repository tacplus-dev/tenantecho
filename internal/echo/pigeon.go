// Copyright (c) 2026 Jody Scott
// SPDX-License-Identifier: Apache-2.0

package echo

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/tacplus-dev/tenantecho/tacplus/chassis/servicemanager"
	eb_pb "github.com/tacplus-dev/tenantecho/tacplus/types/generated/eventbus_support"
	coop_pb "github.com/tacplus-dev/tenantecho/tacplus/types/services/coop"
	pb "github.com/tacplus-dev/tenantecho/tacplus/types/services/tenantecho"
	pigeon_pb "github.com/tacplus-dev/tenantecho/tacplus/types/shared/pigeon"
)

// pigeonKind is the coop kind this proof asks for; the fleet's coop config
// defines it with a mock result.
const pigeonKind = "echo.ping"

// Coop is the part of the generated coop client PingPigeon uses.
type Coop interface {
	CreatePigeon(ctx context.Context, req *coop_pb.PigeonRequest, opts ...grpc.CallOption) (*coop_pb.PigeonTicket, error)
	CancelPigeon(ctx context.Context, req *coop_pb.CancelPigeonRequest, opts ...grpc.CallOption) (*coop_pb.CancelPigeonResponse, error)
}

// Pigeons is the dog's half of the thread: the coop client, the generated
// PigeonMessage publisher, and the subscription delivering the thread back
// (OnMessage). main wires it; the tests fake it.
type Pigeons struct {
	Coop    Coop
	Publish func(context.Context, *coop_pb.PigeonMessage) error
	// Namespace is this instance's tenant namespace, from its own token: the
	// thread's store namespace.
	Namespace string
	// Wait bounds how long a ping waits for the pigeon's result.
	Wait time.Duration

	waiters waiters
}

// PingPigeon asks the coop for an echo.ping pigeon, sends the message as its
// task on the thread, and returns the pigeon's result.
func (s *Server) PingPigeon(ctx context.Context, caller *servicemanager.Caller, req *pb.PingPigeonRequest) (*pb.PingPigeonResponse, error) {
	if s.Pigeons == nil {
		return nil, status.Error(codes.Unimplemented, "this instance has no coop")
	}
	p := s.Pigeons
	wait := p.Wait
	if wait == 0 {
		wait = 30 * time.Second
	}
	ticket, err := p.Coop.CreatePigeon(ctx, &coop_pb.PigeonRequest{
		Kind:       pigeonKind,
		Task:       []byte(req.GetMessage()),
		DeadlineMs: time.Now().Add(wait).UnixMilli(),
	})
	if err != nil {
		return nil, status.Errorf(status.Code(err), "coop: %v", err)
	}
	// The run ends when this ping does, whichever way: a proof's pigeons must
	// not pile up against the tenant's live cap. Unless the caller asked to
	// keep it, to watch the coop's deadline end it.
	defer func() {
		if req.GetKeep() {
			return
		}
		endCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := p.Coop.CancelPigeon(endCtx, &coop_pb.CancelPigeonRequest{RunId: ticket.GetRunId(), Reason: coop_pb.DiscardReason_DISCARD_REASON_DONE}); err != nil {
			zap.L().Warn("cancelling pigeon run", zap.String("run", ticket.GetRunId()), zap.Error(err))
		}
	}()

	ch := p.waiters.add(ticket.GetRunId())
	defer p.waiters.remove(ticket.GetRunId())

	body, err := proto.Marshal(&pigeon_pb.ToPigeon{Body: &pigeon_pb.ToPigeon_Task{Task: &pigeon_pb.PigeonTask{
		Kind: pigeonKind, JobId: ticket.GetRunId(), Request: []byte(req.GetMessage()),
	}}})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "task: %v", err)
	}
	if err := p.Publish(ctx, &coop_pb.PigeonMessage{
		RunId: ticket.GetRunId(), ThreadId: ticket.GetThreadId(), Direction: coop_pb.Direction_DIRECTION_TO_PIGEON,
		Kind: pigeonKind, Seq: 1, Body: body, TenantNamespace: p.Namespace, SentMs: time.Now().UnixMilli(),
	}); err != nil {
		return nil, status.Errorf(codes.Unavailable, "publishing the task: %v", err)
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, status.FromContextError(ctx.Err()).Err()
		case <-timer.C:
			return nil, status.Errorf(codes.DeadlineExceeded, "no result from pigeon %s within %s", ticket.GetRunId(), wait)
		case m := <-ch:
			if e := m.GetError(); e != nil {
				return nil, status.Errorf(codes.Unavailable, "coop on run %s: %s %s", ticket.GetRunId(), e.GetCode(), e.GetDetail())
			}
			var from pigeon_pb.FromPigeon
			if err := proto.Unmarshal(m.GetBody(), &from); err != nil {
				return nil, status.Errorf(codes.Internal, "pigeon message: %v", err)
			}
			r := from.GetResult()
			if r == nil {
				continue // Ready, Progress: keep waiting for the result
			}
			if r.GetStatus() != pigeon_pb.ResultStatus_RESULT_STATUS_OK {
				return nil, status.Errorf(codes.Unavailable, "pigeon failed: %s %s", r.GetFailure().GetCode(), r.GetFailure().GetDetail())
			}
			return &pb.PingPigeonResponse{
				RunId: ticket.GetRunId(), PigeonSpiffeId: ticket.GetPigeonSpiffeId(), Result: string(r.GetResult()),
				ServedBy: s.ServedBy, CallerNamespace: callerNamespace(caller),
			}, nil
		}
	}
}

// OnMessage is the subscription to the coop's thread topic: everything not
// from a dog goes to the ping waiting on that run, if this instance has one.
func (p *Pigeons) OnMessage(_ context.Context, m *coop_pb.PigeonMessage, _ *eb_pb.RecordMeta) error {
	if m.GetDirection() == coop_pb.Direction_DIRECTION_TO_PIGEON {
		return nil
	}
	p.waiters.deliver(m)
	return nil
}

func callerNamespace(caller *servicemanager.Caller) string {
	if caller == nil || caller.Token() == nil {
		return ""
	}
	return caller.Token().GetNamespace()
}

// waiters routes thread messages to the ping waiting on a run.
type waiters struct {
	mu sync.Mutex
	m  map[string]chan *coop_pb.PigeonMessage
}

func (w *waiters) add(runID string) chan *coop_pb.PigeonMessage {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.m == nil {
		w.m = map[string]chan *coop_pb.PigeonMessage{}
	}
	ch := make(chan *coop_pb.PigeonMessage, 16)
	w.m[runID] = ch
	return ch
}

func (w *waiters) remove(runID string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.m, runID)
}

func (w *waiters) deliver(m *coop_pb.PigeonMessage) {
	w.mu.Lock()
	ch := w.m[m.GetRunId()]
	w.mu.Unlock()
	if ch == nil {
		return
	}
	select {
	case ch <- m:
	default:
		zap.L().Warn("pigeon thread waiter is full; dropping", zap.String("run", m.GetRunId()), zap.Int64("seq", m.GetSeq()))
	}
}
