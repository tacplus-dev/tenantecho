// Copyright (c) 2026 Jody Scott
// SPDX-License-Identifier: Apache-2.0

// Package echo is the tenantecho handler: it answers every ping with who
// served it and who asked, which is the whole proof.
package echo

import (
	"context"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tacplus-dev/tenantecho/tacplus/chassis/servicemanager"
	sm "github.com/tacplus-dev/tenantecho/tacplus/s2sm"
	fe "github.com/tacplus-dev/tenantecho/tacplus/s2sm/grpc/servers/services/tenantecho"
	pb "github.com/tacplus-dev/tenantecho/tacplus/types/services/tenantecho"
)

// Server implements the generated TenantEcho handler.
type Server struct {
	// ServedBy is this instance's own SPIFFE identity
	// (tenants/<tenant>/services/tenantecho), read from its own token at start.
	ServedBy string

	mu       sync.RWMutex
	greeting string
}

var _ sm.Handler = (*Server)(nil)

// TenantEchoGrpcImpl returns the handler the generated wrapper serves.
func (s *Server) TenantEchoGrpcImpl() fe.TenantEchoHandler { return s }

// TenantEchoConfig receives this instance's Config at start and on change.
func (s *Server) TenantEchoConfig(cfg *pb.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.greeting = cfg.GetGreeting()
	return nil
}

// Shutdown has nothing to drain.
func (s *Server) Shutdown(context.Context) {}

// Ping answers with the message, this instance's identity, the caller's tenant
// namespace and the configured greeting.
func (s *Server) Ping(_ context.Context, caller *servicemanager.Caller, req *pb.PingRequest) (*pb.PingResponse, error) {
	callerNS := ""
	if caller != nil && caller.Token() != nil {
		callerNS = caller.Token().GetNamespace()
	}
	s.mu.RLock()
	greeting := s.greeting
	s.mu.RUnlock()
	return &pb.PingResponse{
		Message:         req.GetMessage(),
		ServedBy:        s.ServedBy,
		CallerNamespace: callerNS,
		Greeting:        greeting,
	}, nil
}

// TenantechoConfigStateConfig is the generated config hook for the appended
// config-state service; it carries the same Config TenantEchoConfig applied.
func (s *Server) TenantechoConfigStateConfig(*pb.Config) error { return nil }

// TenantechoConfigStateGrpcImpl is the generated per-service config state
// surface. It is served by the config store (kernel), not by this service, so
// it is refused here — the same shape every Go fleet service carries today.
func (s *Server) TenantechoConfigStateGrpcImpl() fe.TenantechoConfigStateServiceHandler {
	return configStateNotHosted{}
}

type configStateNotHosted struct{}

func (configStateNotHosted) GetConfig(context.Context, *servicemanager.Caller) (*pb.Config, error) {
	return nil, errConfigStateNotHosted
}

func (configStateNotHosted) SetConfig(context.Context, *servicemanager.Caller, *pb.Config) (*pb.Config, error) {
	return nil, errConfigStateNotHosted
}

func (configStateNotHosted) WatchConfig(*servicemanager.Caller, grpc.ServerStreamingServer[pb.Config]) error {
	return errConfigStateNotHosted
}

var errConfigStateNotHosted = status.Error(codes.Unimplemented,
	"tenantecho's config state is served by the config store, not by tenantecho")
