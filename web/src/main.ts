// Copyright (c) 2026 Jody Scott
// SPDX-License-Identifier: Apache-2.0

// Wires the ping page to its generated W2S clients (model/w2sm: tenantecho.TenantEcho
// and gateway.Mint, declared in proto/clients/echoweb.proto). Same origin as the
// gateway: its routes, and the session the gateway's own page keeps in localStorage.
import { MintTokenProvider, ServiceManager } from '../model/w2sm/index.js';
import { EchoPage, mountEcho } from './echo.js';

const goHome = () => window.location.assign('/');
const sm = new ServiceManager({ gatewayUrl: '' });
const tokens = new MintTokenProvider(sm.grpc().gateway().mint_client(), localStorage, { onSessionEnded: goHome });
const authed = new ServiceManager({ gatewayUrl: '' }, tokens.provider);

const page = new EchoPage(authed.grpc().tenantecho().tenant_echo_client(), tokens, goHome);
if (page.start()) mountEcho(document.getElementById('app') as HTMLElement, page);
