// Copyright (c) 2026 Jody Scott
// SPDX-License-Identifier: Apache-2.0

package echo

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	pb "github.com/tacplus-dev/tenantecho/tacplus/types/services/tenantecho"
	spec "github.com/tacplus-dev/tenantecho/tacplus/types/sys/spec"
)

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
