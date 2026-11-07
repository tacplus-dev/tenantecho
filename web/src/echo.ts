// Copyright (c) 2026 Jody Scott
// SPDX-License-Identifier: Apache-2.0

// The ping page: pings the named tenant's tenantecho instance and shows the
// reply and the tenant it came from. It signs nobody in: it runs on the
// gateway's origin and uses the gateway's session, and a signed-out user is
// sent to the gateway's page at /.
import { create } from '@bufbuild/protobuf';
import { W2SError } from '../model/w2sm/index.js';
import { PingRequestSchema, type PingRequest, type PingResponse } from '../model/types/services/tenantecho/services/index.js';

/** The generated W2S TenantEcho client's ping, as the page uses it. */
export interface PingClient {
  ping(request: PingRequest): Promise<PingResponse>;
}

/** The gateway's session, as the generated MintTokenProvider holds it. */
export interface Session {
  hasRefreshToken(): boolean;
  setTenantNamespace(ns: string | null): void;
}

export type PingResult =
  | { ok: true; message: string; servedBy: string; tenant: string; callerNamespace: string; greeting: string }
  | { ok: false; error: string };

/** A tenant's tid: 8 characters of [0-9a-z]. */
const TID = /^[0-9a-z]{8}$/;

// gRPC codes the page tells apart.
const INVALID_ARGUMENT = 3;
const PERMISSION_DENIED = 7;
const UNAUTHENTICATED = 16;

/** The tenant in a SPIFFE id or namespace (…/tenants/<tid>/…); '' when none. */
export function tenantOf(idOrNamespace: string): string {
  return /\/tenants\/([^/]+)/.exec(idOrNamespace)?.[1] ?? '';
}

export class EchoPage {
  constructor(
    private readonly client: PingClient,
    private readonly session: Session,
    private readonly goHome: () => void,
  ) {}

  /** False, having sent the user to the gateway, when nobody is signed in. */
  start(): boolean {
    if (this.session.hasRefreshToken()) return true;
    this.goHome();
    return false;
  }

  /**
   * Pings tenant `tid`: its tid is the route field, and the call carries a
   * token minted for that tenant's namespace, so only its members get in.
   */
  async ping(tid: string, message: string): Promise<PingResult> {
    tid = tid.trim();
    if (!TID.test(tid)) return { ok: false, error: 'A tenant is its tid: 8 characters of 0-9 and a-z.' };
    if (!message) return { ok: false, error: 'Type a message to send.' };
    try {
      this.session.setTenantNamespace(`/tenants/${tid}`);
      const resp = await this.client.ping(create(PingRequestSchema, { tenant: tid, message }));
      return {
        ok: true,
        message: resp.message,
        servedBy: resp.servedBy,
        tenant: tenantOf(resp.servedBy),
        callerNamespace: resp.callerNamespace,
        greeting: resp.greeting,
      };
    } catch (e) {
      if (e instanceof W2SError) {
        switch (e.code) {
          case UNAUTHENTICATED:
            this.goHome();
            return { ok: false, error: 'Your session has ended: sign in again.' };
          case PERMISSION_DENIED:
            return { ok: false, error: `Refused: you are not a member of tenant ${tid}.` };
          case INVALID_ARGUMENT:
            return { ok: false, error: `Refused as invalid: ${e.message}` };
        }
      }
      return { ok: false, error: e instanceof Error ? e.message : String(e) };
    }
  }
}

/** Renders the page's form into `root`; each submit pings and shows the result. */
export function mountEcho(root: HTMLElement, page: EchoPage): void {
  const doc = root.ownerDocument;
  const form = doc.createElement('form');
  const heading = doc.createElement('h1');
  heading.textContent = 'Echo';
  const tenant = input(doc, 'tenant', 'Tenant (tid)');
  const message = input(doc, 'message', 'Message');
  message.input.value = 'ping';
  const send = doc.createElement('button');
  send.type = 'submit';
  send.textContent = 'Ping';
  const output = doc.createElement('output');
  form.append(heading, tenant.label, message.label, send, output);
  root.replaceChildren(form);

  form.addEventListener('submit', async (ev) => {
    ev.preventDefault();
    send.disabled = true;
    output.replaceChildren();
    try {
      output.replaceChildren(render(doc, await page.ping(tenant.input.value, message.input.value)));
    } finally {
      send.disabled = false;
    }
  });
}

function input(doc: Document, name: string, text: string): { label: HTMLLabelElement; input: HTMLInputElement } {
  const label = doc.createElement('label');
  const el = doc.createElement('input');
  el.name = name;
  el.required = true;
  label.append(`${text} `, el);
  return { label, input: el };
}

function render(doc: Document, result: PingResult): HTMLElement {
  if (!result.ok) {
    const p = doc.createElement('p');
    p.className = 'error';
    p.textContent = result.error;
    return p;
  }
  const dl = doc.createElement('dl');
  const rows: [string, string][] = [
    ['Reply', result.message],
    ['From tenant', result.tenant || '(none)'],
    ['Served by', result.servedBy],
    ['Your tenant', result.callerNamespace],
    ['Greeting', result.greeting],
  ];
  for (const [k, v] of rows) {
    const dt = doc.createElement('dt');
    dt.textContent = k;
    const dd = doc.createElement('dd');
    dd.textContent = v;
    dl.append(dt, dd);
  }
  return dl;
}
