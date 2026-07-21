// Minimal Cloudflare Workers types used by this shim. The full set is available
// via @cloudflare/workers-types or `wrangler types`; we declare only what this
// tiny worker uses, to avoid a heavyweight dev dependency.
//
// Note: ForwardableEmailMessage has a forward() method, but this shim does not
// use it (there is no archive mailbox) — so it is intentionally omitted here.

export interface ForwardableEmailMessage {
  readonly from: string;
  readonly to: string;
  readonly raw: ReadableStream<Uint8Array>;
  readonly rawSize: number;
}

export interface ExecutionContext {
  waitUntil(promise: Promise<unknown>): void;
  passThroughOnException(): void;
}

export interface ExportedHandler<Env = unknown> {
  email?(
    message: ForwardableEmailMessage,
    env: Env,
    ctx: ExecutionContext
  ): Promise<void>;
}
