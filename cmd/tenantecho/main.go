// Copyright (c) 2026 Jody Scott
// SPDX-License-Identifier: Apache-2.0

// tenantecho is the smallest tenant-owned fleet service: one instance per
// tenant namespace, answering pings with who served them. It exists to prove
// the gateway routes each tenant's call to that tenant's instance.
package main

import (
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/tacplus-dev/tenantecho/internal/echo"
	loggerchassis "github.com/tacplus-dev/tenantecho/tacplus/chassis/logger"
	"github.com/tacplus-dev/tenantecho/tacplus/chassis/token"
	sm "github.com/tacplus-dev/tenantecho/tacplus/s2sm"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	logger, err := loggerchassis.InstallFromConfig(nil)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync() //nolint:errcheck

	s, err := sm.New()
	if err != nil {
		return fmt.Errorf("servicemanager: %w", err)
	}
	defer s.Close()

	// This instance's own identity, from the token the fleet minted for it:
	// the SPIFFE id spiffe://<td>/tenants/<tenant>/services/tenantecho for a
	// tenant-owned instance. The service path alone is "tenantecho" for every
	// tenant's instance, which is exactly what the proof must distinguish.
	raw, err := s.GetToken()
	if err != nil {
		return fmt.Errorf("own token: %w", err)
	}
	own, err := token.ParseToken(raw)
	if err != nil {
		return fmt.Errorf("own token: %w", err)
	}
	servedBy := own.GetSubject()
	if servedBy == "" {
		return errors.New("own token has no subject")
	}
	pigeons := &echo.Pigeons{
		Coop:      s.Grpc().Clients().Services().CoopClient(),
		Publish:   s.CreateCoopPigeonMessageBusStreamPublisher(),
		Namespace: own.GetNamespace(),
		Self:      servedBy,
	}
	// The coop's thread comes back on this instance's subscription.
	s.SetOnStreamCoopPigeonMessageBusStreamFunc(pigeons.OnMessage)
	zap.L().Info("tenantecho", zap.String("served_by", servedBy))
	return s.Run(&echo.Server{ServedBy: servedBy, Pigeons: pigeons})
}
