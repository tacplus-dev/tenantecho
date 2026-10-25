// Copyright (c) 2026 Jody Scott
// SPDX-License-Identifier: Apache-2.0

package echo

import (
	"context"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	coop_pb "github.com/tacplus-dev/tenantecho/tacplus/types/services/coop"
	pb "github.com/tacplus-dev/tenantecho/tacplus/types/services/tenantecho"
	pigeon_pb "github.com/tacplus-dev/tenantecho/tacplus/types/shared/pigeon"
)

// fakeCoop is the coop as a dog sees it: a ticket, then the pigeon's messages
// on the thread. reply is what the fake pigeon puts on the thread after the
// dog's task arrives, in order.
type fakeCoop struct {
	created   []*coop_pb.PigeonRequest
	cancelled []string
	reply     []*coop_pb.PigeonMessage
}

func (f *fakeCoop) CreatePigeon(_ context.Context, req *coop_pb.PigeonRequest, _ ...grpc.CallOption) (*coop_pb.PigeonTicket, error) {
	f.created = append(f.created, req)
	return &coop_pb.PigeonTicket{RunId: "run1", ThreadId: "run1", PigeonSpiffeId: "spiffe://local.dev/system/services/coop"}, nil
}
func (f *fakeCoop) CancelPigeon(_ context.Context, req *coop_pb.CancelPigeonRequest, _ ...grpc.CallOption) (*coop_pb.CancelPigeonResponse, error) {
	if req.GetReason() == coop_pb.DiscardReason_DISCARD_REASON_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "reason must be a defined value") // what the generated validation says
	}
	f.cancelled = append(f.cancelled, req.GetRunId())
	return &coop_pb.CancelPigeonResponse{}, nil
}

func fromPigeon(t *testing.T, body *pigeon_pb.FromPigeon) *coop_pb.PigeonMessage {
	t.Helper()
	raw, err := proto.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return &coop_pb.PigeonMessage{RunId: "run1", ThreadId: "run1", Direction: coop_pb.Direction_DIRECTION_FROM_PIGEON, Body: raw}
}

func newDog(t *testing.T, coop *fakeCoop) (*Server, *[]*coop_pb.PigeonMessage) {
	t.Helper()
	var sent []*coop_pb.PigeonMessage
	p := &Pigeons{Coop: coop, Namespace: "/tenants/acme", Wait: 2 * time.Second}
	// Publishing the task is what makes the (fake) pigeon answer, as the real
	// thread would deliver it back through the subscription.
	p.Publish = func(ctx context.Context, m *coop_pb.PigeonMessage) error {
		sent = append(sent, m)
		go func() {
			for _, r := range coop.reply {
				_ = p.OnMessage(ctx, r, nil)
			}
		}()
		return nil
	}
	return &Server{ServedBy: "spiffe://local.dev/tenants/acme/services/tenantecho", Pigeons: p}, &sent
}

func TestPingPigeonRoundTrip(t *testing.T) {
	coop := &fakeCoop{}
	coop.reply = []*coop_pb.PigeonMessage{
		// A TO_PIGEON echo on the thread (the dog's own task) is never a reply.
		{RunId: "run1", Direction: coop_pb.Direction_DIRECTION_TO_PIGEON},
		fromPigeon(t, &pigeon_pb.FromPigeon{Body: &pigeon_pb.FromPigeon_Ready{Ready: &pigeon_pb.Ready{}}}),
		fromPigeon(t, &pigeon_pb.FromPigeon{Body: &pigeon_pb.FromPigeon_Result{Result: &pigeon_pb.PigeonResult{
			Status: pigeon_pb.ResultStatus_RESULT_STATUS_OK, Result: []byte("pong"),
		}}}),
	}
	s, sent := newDog(t, coop)
	resp, err := s.PingPigeon(context.Background(), nil, &pb.PingPigeonRequest{Message: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetResult() != "pong" || resp.GetRunId() != "run1" || resp.GetPigeonSpiffeId() != "spiffe://local.dev/system/services/coop" ||
		resp.GetServedBy() != "spiffe://local.dev/tenants/acme/services/tenantecho" {
		t.Fatalf("resp %+v", resp)
	}
	if len(coop.created) != 1 || coop.created[0].GetKind() != "echo.ping" || string(coop.created[0].GetTask()) != "hello" || coop.created[0].GetDeadlineMs() == 0 {
		t.Fatalf("created %+v", coop.created)
	}
	if len(*sent) != 1 {
		t.Fatalf("sent %d messages", len(*sent))
	}
	m := (*sent)[0]
	var to pigeon_pb.ToPigeon
	if err := proto.Unmarshal(m.GetBody(), &to); err != nil {
		t.Fatal(err)
	}
	if m.GetDirection() != coop_pb.Direction_DIRECTION_TO_PIGEON || m.GetRunId() != "run1" || m.GetTenantNamespace() != "/tenants/acme" ||
		m.GetSpiffeId() != "" || // eventbus stamps the sender from the publish token; the dog sends none
		to.GetTask().GetKind() != "echo.ping" || string(to.GetTask().GetRequest()) != "hello" {
		t.Fatalf("task message %+v / %+v", m, &to)
	}
	// The run is cancelled once the ping has its answer.
	if len(coop.cancelled) != 1 || coop.cancelled[0] != "run1" {
		t.Fatalf("cancelled %v", coop.cancelled)
	}
}

// keep leaves the run to the coop's deadline: no cancel after the result.
func TestPingPigeonKeepDoesNotCancel(t *testing.T) {
	coop := &fakeCoop{reply: []*coop_pb.PigeonMessage{fromPigeon(t, &pigeon_pb.FromPigeon{Body: &pigeon_pb.FromPigeon_Result{Result: &pigeon_pb.PigeonResult{
		Status: pigeon_pb.ResultStatus_RESULT_STATUS_OK, Result: []byte("pong"),
	}}})}}
	s, _ := newDog(t, coop)
	resp, err := s.PingPigeon(context.Background(), nil, &pb.PingPigeonRequest{Message: "x", Keep: true})
	if err != nil || resp.GetResult() != "pong" {
		t.Fatalf("%v %v", resp, err)
	}
	if len(coop.cancelled) != 0 {
		t.Fatalf("keep must leave the run alive: cancelled %v", coop.cancelled)
	}
}

func TestPingPigeonFailuresAndTimeout(t *testing.T) {
	coop := &fakeCoop{reply: []*coop_pb.PigeonMessage{fromPigeon(t, &pigeon_pb.FromPigeon{Body: &pigeon_pb.FromPigeon_Result{Result: &pigeon_pb.PigeonResult{
		Status: pigeon_pb.ResultStatus_RESULT_STATUS_FAILED, Failure: &pigeon_pb.Failure{Code: pigeon_pb.FailureCode_FAILURE_CODE_INVALID_TASK, Detail: "bad"},
	}}})}}
	s, _ := newDog(t, coop)
	_, err := s.PingPigeon(context.Background(), nil, &pb.PingPigeonRequest{Message: "x"})
	if status.Code(err) != codes.Unavailable || !strings.Contains(err.Error(), "INVALID_TASK") {
		t.Fatalf("failed result: %v", err)
	}

	coop = &fakeCoop{reply: []*coop_pb.PigeonMessage{{RunId: "run1", Direction: coop_pb.Direction_DIRECTION_FROM_COOP,
		Error: &coop_pb.MessageError{Code: coop_pb.MessageErrorCode_MESSAGE_ERROR_CODE_PIGEON_GONE, Detail: "expired"}}}}
	s, _ = newDog(t, coop)
	_, err = s.PingPigeon(context.Background(), nil, &pb.PingPigeonRequest{Message: "x"})
	if status.Code(err) != codes.Unavailable || !strings.Contains(err.Error(), "PIGEON_GONE expired") {
		t.Fatalf("gone: %v", err)
	}

	coop = &fakeCoop{} // a pigeon that never answers
	s, _ = newDog(t, coop)
	s.Pigeons.Wait = 50 * time.Millisecond
	_, err = s.PingPigeon(context.Background(), nil, &pb.PingPigeonRequest{Message: "x"})
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("timeout: %v", err)
	}
	if len(coop.cancelled) != 1 {
		t.Fatalf("a timed-out run is still cancelled: %v", coop.cancelled)
	}

	if _, err := (&Server{}).PingPigeon(context.Background(), nil, &pb.PingPigeonRequest{Message: "x"}); status.Code(err) != codes.Unimplemented {
		t.Fatalf("no coop: %v", err)
	}
}
