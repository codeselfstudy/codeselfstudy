// Testable helpers for the email shim. fetch and sleep are injected so the whole
// orchestration can be unit-tested with `bun test`; index.ts is only the runtime
// wiring.

import type { ForwardableEmailMessage } from "./cf-types";

export interface IngestOptions {
  token: string;
  from: string;
  to: string;
}

/** buildHeaders returns the headers for an /api/ingest POST. */
export function buildHeaders(opts: IngestOptions): Record<string, string> {
  return {
    Authorization: `Bearer ${opts.token}`,
    "Content-Type": "message/rfc822",
    "X-Envelope-From": opts.from,
    "X-Envelope-To": opts.to,
  };
}

export interface RetryConfig {
  /** Total number of attempts, including the first (default 2). */
  attempts?: number;
  /** Delay between attempts in ms (default 5000 — covers a Fly cold start). */
  delayMs?: number;
  /** Injectable sleep, for tests. */
  sleep?: (ms: number) => Promise<void>;
}

/**
 * postWithRetry POSTs the raw email to the ingest endpoint, retrying on non-2xx
 * responses and network errors. It resolves with the successful Response or
 * rejects after the last attempt fails.
 */
export async function postWithRetry(
  fetchFn: typeof fetch,
  url: string,
  body: ArrayBuffer | Uint8Array,
  opts: IngestOptions,
  cfg: RetryConfig = {}
): Promise<Response> {
  const attempts = cfg.attempts ?? 2;
  const delayMs = cfg.delayMs ?? 5000;
  const sleep =
    cfg.sleep ?? ((ms: number) => new Promise((r) => setTimeout(r, ms)));
  const headers = buildHeaders(opts);

  let lastError: unknown;
  for (let attempt = 1; attempt <= attempts; attempt++) {
    try {
      const resp = await fetchFn(url, { method: "POST", headers, body });
      if (resp.ok) return resp;
      lastError = new Error(`ingest returned HTTP ${resp.status}`);
      // Discard the body so the connection isn't held open across the delay.
      await resp.body?.cancel();
    } catch (err) {
      lastError = err;
    }
    if (attempt < attempts) await sleep(delayMs);
  }
  throw lastError instanceof Error ? lastError : new Error(String(lastError));
}

/** Env is the subset of Worker bindings the handler needs. */
export interface Env {
  /** URL of the Go app's /api/ingest endpoint. */
  INGEST_URL: string;
  /** Shared bearer token, matching the Go app's INGEST_TOKEN. */
  INGEST_TOKEN: string;
}

// Cloudflare Email Routing caps messages at 25 MB; refuse the POST above this to
// stay well within the Worker's memory. There is no archive copy, so an oversize
// message is bounced (see handleEmail) rather than silently dropped.
export const MAX_RAW_BYTES = 20 * 1024 * 1024;

/**
 * handleEmail POSTs an incoming message to the Go app's /api/ingest.
 *
 * There is no archive mailbox, so a thrown error IS the backstop: Cloudflare
 * bounces the message and the sender re-delivers, and nothing is silently lost.
 * A persistent POST failure therefore propagates out of postWithRetry (not
 * swallowed), and an oversize message is refused with a throw before the POST.
 */
export async function handleEmail(
  message: ForwardableEmailMessage,
  env: Env,
  fetchFn: typeof fetch,
  cfg: RetryConfig = {}
): Promise<void> {
  if (message.rawSize > MAX_RAW_BYTES) {
    throw new Error(
      `raw size ${message.rawSize} exceeds ${MAX_RAW_BYTES}; refusing to POST`
    );
  }

  const raw = await new Response(message.raw).arrayBuffer();
  await postWithRetry(
    fetchFn,
    env.INGEST_URL,
    raw,
    { token: env.INGEST_TOKEN, from: message.from, to: message.to },
    cfg
  );
}
