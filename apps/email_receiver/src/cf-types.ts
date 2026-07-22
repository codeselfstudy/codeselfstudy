// Minimal Cloudflare Workers types used by this shim. The full set is available
// via @cloudflare/workers-types or `wrangler types`; we declare only what this
// tiny worker uses, to avoid a heavyweight dev dependency.
//
// Note: ForwardableEmailMessage also has forward() and reply() methods; this
// shim uses neither (there is no archive mailbox), so they are omitted. It does
// use setReject() to permanently reject a message it will never accept.

export interface ForwardableEmailMessage {
  readonly from: string;
  readonly to: string;
  readonly raw: ReadableStream<Uint8Array>;
  readonly rawSize: number;
  /** Reject the message with an SMTP error (a permanent, non-retryable bounce). */
  setReject(reason: string): void;
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
