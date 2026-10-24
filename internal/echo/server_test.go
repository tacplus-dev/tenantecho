// Copyright (c) 2026 Jody Scott
// SPDX-License-Identifier: Apache-2.0

package echo

import (
	"context"
	"testing"

	"github.com/tacplus-dev/tenantecho/tacplus/chassis/servicemanager"
	"github.com/tacplus-dev/tenantecho/tacplus/chassis/token"
	pb "github.com/tacplus-dev/tenantecho/tacplus/types/services/tenantecho"
	tokenProto "github.com/tacplus-dev/tenantecho/tacplus/types/sys/token"
)

// A ping names the instance that answered, the caller's tenant and the
// instance's configured greeting; a misrouted call shows as served_by naming
// a tenant other than caller_namespace.
func TestPingNamesServerAndCaller(t *testing.T) {
	s := &Server{ServedBy: "spiffe://local.dev/tenants/acme/services/tenantecho"}
	if err := s.TenantEchoConfig(&pb.Config{Greeting: "ahoy"}); err != nil {
		t.Fatal(err)
	}
	ns := tokenProto.JwtPayloadTypeToString(tokenProto.JwtPayloadType_JWT_PAYLOAD_TYPE_NAMESPACE)
	caller := servicemanager.NewCaller(token.NewImmutableToken(&tokenProto.Token{Payload: map[string]string{ns: "acme"}}), "req-1")

	resp, err := s.Ping(context.Background(), caller, &pb.PingRequest{Message: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetMessage() != "hello" || resp.GetServedBy() != "spiffe://local.dev/tenants/acme/services/tenantecho" ||
		resp.GetCallerNamespace() != "acme" || resp.GetGreeting() != "ahoy" {
		t.Fatalf("resp = %+v", resp)
	}

	// Config changes take effect on the next ping.
	if err := s.TenantEchoConfig(&pb.Config{Greeting: "yo"}); err != nil {
		t.Fatal(err)
	}
	if resp, _ := s.Ping(context.Background(), nil, &pb.PingRequest{Message: "x"}); resp.GetGreeting() != "yo" || resp.GetCallerNamespace() != "" {
		t.Fatalf("after config change: %+v", resp)
	}
}
