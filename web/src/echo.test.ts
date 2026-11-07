// Copyright (c) 2026 Jody Scott
// SPDX-License-Identifier: Apache-2.0
import { create } from '@bufbuild/protobuf';
import { beforeEach, describe, expect, it } from 'vitest';
import { W2SError } from '../model/w2sm/index.js';
import { PingResponseSchema, type PingRequest, type PingResponse } from '../model/types/services/tenantecho/services/index.js';
import { EchoPage, mountEcho, tenantOf, type PingClient, type Session } from './echo.js';

const TID = 'a1b2c3d4';

class FakeClient implements PingClient {
  requests: PingRequest[] = [];
  reply: PingResponse | Error = create(PingResponseSchema, {
    message: 'hello',
    servedBy: `spiffe://local.dev/tenants/${TID}/services/tenantecho`,
    callerNamespace: `/tenants/${TID}`,
    greeting: 'hi from a1',
  });
  async ping(request: PingRequest): Promise<PingResponse> {
    this.requests.push(request);
    if (this.reply instanceof Error) throw this.reply;
    return this.reply;
  }
}

class FakeSession implements Session {
  signedIn = true;
  namespaces: (string | null)[] = [];
  hasRefreshToken(): boolean { return this.signedIn; }
  setTenantNamespace(ns: string | null): void { this.namespaces.push(ns); }
}

let client: FakeClient;
let session: FakeSession;
let home: number;
let page: EchoPage;

beforeEach(() => {
  client = new FakeClient();
  session = new FakeSession();
  home = 0;
  page = new EchoPage(client, session, () => { home++; });
});

describe('tenantOf', () => {
  it('reads the tenant from a SPIFFE id or a namespace', () => {
    expect(tenantOf(`spiffe://local.dev/tenants/${TID}/services/tenantecho`)).toBe(TID);
    expect(tenantOf(`/tenants/${TID}`)).toBe(TID);
    expect(tenantOf('spiffe://local.dev/services/tenantecho')).toBe('');
    expect(tenantOf('')).toBe('');
  });
});

describe('EchoPage', () => {
  it('sends a signed-out user to the gateway', () => {
    session.signedIn = false;
    expect(page.start()).toBe(false);
    expect(home).toBe(1);
  });

  it('stays for a signed-in user', () => {
    expect(page.start()).toBe(true);
    expect(home).toBe(0);
  });

  it('pings the named tenant with a token minted for it, and shows where the reply came from', async () => {
    const result = await page.ping(` ${TID} `, 'hello');
    expect(session.namespaces).toEqual([`/tenants/${TID}`]);
    expect(client.requests).toHaveLength(1);
    expect(client.requests[0].tenant).toBe(TID);
    expect(client.requests[0].message).toBe('hello');
    expect(result).toEqual({
      ok: true,
      message: 'hello',
      servedBy: `spiffe://local.dev/tenants/${TID}/services/tenantecho`,
      tenant: TID,
      callerNamespace: `/tenants/${TID}`,
      greeting: 'hi from a1',
    });
  });

  it('refuses a malformed tid or an empty message without calling', async () => {
    expect(await page.ping('acme', 'hello')).toEqual({ ok: false, error: expect.stringContaining('8 characters') });
    expect(await page.ping(TID, '')).toEqual({ ok: false, error: expect.stringContaining('message') });
    expect(client.requests).toHaveLength(0);
  });

  it('says a refused tenant is not the caller\'s', async () => {
    client.reply = new W2SError(7, 'gRPC-over-REST error: 403 Forbidden', 403);
    const result = await page.ping(TID, 'hello');
    expect(result).toEqual({ ok: false, error: expect.stringContaining('not a member') });
    expect(home).toBe(0);
  });

  it('sends the user to the gateway when the session has ended', async () => {
    client.reply = new W2SError(16, 'w2sm: no refresh token — sign in first');
    const result = await page.ping(TID, 'hello');
    expect(result.ok).toBe(false);
    expect(home).toBe(1);
  });

  it('reports any other failure', async () => {
    client.reply = new Error('network down');
    expect(await page.ping(TID, 'hello')).toEqual({ ok: false, error: 'network down' });
  });
});

describe('mountEcho', () => {
  it('pings from the form and shows the reply and its tenant', async () => {
    const root = document.createElement('main');
    mountEcho(root, page);
    (root.querySelector('input[name=tenant]') as HTMLInputElement).value = TID;
    (root.querySelector('input[name=message]') as HTMLInputElement).value = 'hello';
    (root.querySelector('form') as HTMLFormElement).requestSubmit();
    await vi_waitFor(() => root.querySelector('dl') !== null);
    const text = root.querySelector('output')!.textContent!;
    expect(text).toContain('hello');
    expect(text).toContain(TID);
    expect(text).toContain('hi from a1');
    expect(client.requests[0].tenant).toBe(TID);
  });

  it('shows an error', async () => {
    client.reply = new W2SError(7, 'forbidden', 403);
    const root = document.createElement('main');
    mountEcho(root, page);
    (root.querySelector('input[name=tenant]') as HTMLInputElement).value = TID;
    (root.querySelector('input[name=message]') as HTMLInputElement).value = 'hello';
    (root.querySelector('form') as HTMLFormElement).requestSubmit();
    await vi_waitFor(() => root.querySelector('.error') !== null);
    expect(root.querySelector('.error')!.textContent).toContain('not a member');
  });
});

async function vi_waitFor(cond: () => boolean): Promise<void> {
  for (let i = 0; i < 100 && !cond(); i++) await new Promise((r) => setTimeout(r, 0));
  expect(cond()).toBe(true);
}
