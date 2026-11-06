// Copyright (c) 2026 Jody Scott
// SPDX-License-Identifier: Apache-2.0

package echo

import (
	"context"
	"slices"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/tacplus-dev/tenantecho/tacplus/chassis/grpcserver"
	"github.com/tacplus-dev/tenantecho/tacplus/chassis/serviceconnection"
	"github.com/tacplus-dev/tenantecho/tacplus/chassis/token"
	fe "github.com/tacplus-dev/tenantecho/tacplus/s2sm/grpc/servers/services/tenantecho"
	pb "github.com/tacplus-dev/tenantecho/tacplus/types/services/tenantecho"
	spec "github.com/tacplus-dev/tenantecho/tacplus/types/sys/spec"
	tokenProto "github.com/tacplus-dev/tenantecho/tacplus/types/sys/token"
)

// instance wires one tenant's tenantecho the way the generated s2sm does
// (route fields registered, CheckRoute as the backstop), with tid in its
// subject. Authentication is stubbed: this is about routing, not the door.
func instance(t *testing.T, name, tid string) *fe.TenantEcho {
	t.Helper()
	auth, err := grpcserver.NewAuthorizer(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	fe.RegisterTenantEchoMethodAuth(auth)
	served := "spiffe://local.dev/tenants/" + tid + "/services/tenantecho"
	auth.SetHostIdentity("/tenants/"+name, served)

	w := fe.NewTenantEcho(&Server{ServedBy: served})
	authn := func(context.Context) (*token.TokenWrapper, error) {
		return token.NewImmutableToken(&tokenProto.Token{}).SetNamespace("/tenants/" + name), nil
	}
	w.SetAuthenticateFunc(authn)
	w.SetAuthorizeFunc(func(ctx context.Context, _ string, _ any) (*token.TokenWrapper, error) { return authn(ctx) })
	w.SetAuthenticateAtOrBelowFunc(func(ctx context.Context, _ string) (*token.TokenWrapper, error) { return authn(ctx) })
	w.SetRouteCheckFunc(auth.CheckRoute)
	w.SetRouteNamespaceFunc(auth.RouteNamespace)
	return w
}

// A call reaches the instance of the tenant its request names, and only that
// one: the caller picks the instance by the route field (as the gateway does,
// serviceconnection.RouteTenant), and each instance refuses a call naming
// another tenant. An empty or malformed tid is refused by the caller before
// anything is dialled and, should it arrive anyway, by the instance.
func TestCallReachesTheNamedTenantsInstance(t *testing.T) {
	const acme, beta = "acme0001", "beta0002"
	instances := map[string]*fe.TenantEcho{acme: instance(t, "acme", acme), beta: instance(t, "beta", beta)}
	ctx := context.Background()

	for _, tid := range []string{acme, beta} {
		target, err := serviceconnection.RouteTenant("tenant", tid)
		if err != nil {
			t.Fatalf("RouteTenant(%q): %v", tid, err)
		}
		resp, err := instances[target].Ping(ctx, &pb.PingRequest{Message: "hi", Tenant: tid})
		if err != nil {
			t.Fatalf("Ping for %s: %v", tid, err)
		}
		if want := "spiffe://local.dev/tenants/" + tid + "/services/tenantecho"; resp.GetServedBy() != want {
			t.Fatalf("Ping for %s served by %q, want %q", tid, resp.GetServedBy(), want)
		}
	}

	// Another tenant's instance refuses, on every rpc.
	if _, err := instances[beta].Ping(ctx, &pb.PingRequest{Message: "hi", Tenant: acme}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("Ping for %s on %s's instance: want PermissionDenied, got %v", acme, beta, err)
	}
	if _, err := instances[beta].PingPigeon(ctx, &pb.PingPigeonRequest{Message: "hi", Tenant: acme}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("PingPigeon for %s on %s's instance: want PermissionDenied, got %v", acme, beta, err)
	}

	// A missing or malformed tid is never defaulted: refused on both sides.
	for _, bad := range []string{"", "acme", "ACME0001", "acme0001 "} {
		if _, err := serviceconnection.RouteTenant("tenant", bad); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("caller with tenant %q: want InvalidArgument, got %v", bad, err)
		}
		if _, err := instances[acme].Ping(ctx, &pb.PingRequest{Message: "hi", Tenant: bad}); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("Ping with tenant %q: want InvalidArgument, got %v", bad, err)
		}
		if _, err := instances[acme].PingPigeon(ctx, &pb.PingPigeonRequest{Message: "hi", Tenant: bad}); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("PingPigeon with tenant %q: want InvalidArgument, got %v", bad, err)
		}
	}
}

// Every TenantEcho rpc routes by its request's tenant (the target tid), a
// plain top-level string, through the default door: only members of the
// target tenant get in, so a call lands in the caller's own tenant's instance
// or is refused. That is the proof; RPC_FLAG_ANY_TENANT would void it.
func TestEveryRpcRoutesByTenant(t *testing.T) {
	svc := pb.File_services_tenantecho_proto.Services().ByName("TenantEcho")
	if svc == nil {
		t.Fatal("no TenantEcho service")
	}
	methods := svc.Methods()
	if methods.Len() == 0 {
		t.Fatal("TenantEcho has no rpcs")
	}
	for i := range methods.Len() {
		m := methods.Get(i)
		rpc, _ := proto.GetExtension(m.Options(), spec.E_Rpc).(*spec.RpcAnnotation)
		if got := rpc.GetRouteField(); got != "tenant" {
			t.Errorf("%s: route_field = %q, want \"tenant\"", m.Name(), got)
			continue
		}
		if slices.Contains(rpc.GetFlags(), spec.RpcFlag_RPC_FLAG_ANY_TENANT) {
			t.Errorf("%s: RPC_FLAG_ANY_TENANT opens the tenant door to other tenants", m.Name())
		}
		f := m.Input().Fields().ByName("tenant")
		if f == nil || f.Kind() != protoreflect.StringKind || f.Cardinality() == protoreflect.Repeated || f.HasPresence() {
			t.Errorf("%s: %s.tenant must be a plain string field", m.Name(), m.Input().Name())
		}
	}
}
