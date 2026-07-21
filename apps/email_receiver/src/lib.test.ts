import { describe, expect, test } from "bun:test";
import {
  buildHeaders,
  handleEmail,
  MAX_RAW_BYTES,
  postWithRetry,
  type Env,
  type IngestOptions,
} from "./lib";
import type { ForwardableEmailMessage } from "./cf-types";

const opts: IngestOptions = { token: "tok", from: "sender@x", to: "deals@y" };
const noSleep = () => Promise.resolve();

describe("buildHeaders", () => {
  test("sets auth, content-type, and envelope headers", () => {
    const h = buildHeaders(opts);
    expect(h["Authorization"]).toBe("Bearer tok");
    expect(h["Content-Type"]).toBe("message/rfc822");
    expect(h["X-Envelope-From"]).toBe("sender@x");
    expect(h["X-Envelope-To"]).toBe("deals@y");
  });
});

describe("postWithRetry", () => {
  test("returns on first success (defaults, no sleep)", async () => {
    let calls = 0;
    const fetchFn = (async () => {
      calls++;
      return new Response("ok");
    }) as unknown as typeof fetch;
    const resp = await postWithRetry(
      fetchFn,
      "http://x/ingest",
      new Uint8Array([1]),
      opts
    );
    expect(resp.status).toBe(200);
    expect(calls).toBe(1);
  });

  test("retries once then succeeds", async () => {
    let calls = 0;
    const fetchFn = (async () => {
      calls++;
      return calls === 1
        ? new Response("", { status: 500 })
        : new Response("ok");
    }) as unknown as typeof fetch;
    const resp = await postWithRetry(
      fetchFn,
      "http://x/ingest",
      new Uint8Array([1]),
      opts,
      {
        attempts: 2,
        sleep: noSleep,
      }
    );
    expect(resp.status).toBe(200);
    expect(calls).toBe(2);
  });

  test("throws after all attempts return non-2xx", async () => {
    let calls = 0;
    const fetchFn = (async () => {
      calls++;
      return new Response("", { status: 503 });
    }) as unknown as typeof fetch;
    await expect(
      postWithRetry(fetchFn, "http://x/ingest", new Uint8Array([1]), opts, {
        attempts: 2,
        sleep: noSleep,
      })
    ).rejects.toThrow("HTTP 503");
    expect(calls).toBe(2);
  });

  test("throws after network errors on every attempt", async () => {
    let calls = 0;
    const fetchFn = (async () => {
      calls++;
      throw new Error("network down");
    }) as unknown as typeof fetch;
    await expect(
      postWithRetry(fetchFn, "http://x/ingest", new Uint8Array([1]), opts, {
        attempts: 2,
        sleep: noSleep,
      })
    ).rejects.toThrow("network down");
    expect(calls).toBe(2);
  });

  test("sends POST with headers and body", async () => {
    let seenUrl = "";
    let seenInit: RequestInit | undefined;
    const fetchFn = (async (url: string, init: RequestInit) => {
      seenUrl = url;
      seenInit = init;
      return new Response("ok");
    }) as unknown as typeof fetch;
    await postWithRetry(
      fetchFn,
      "http://x/ingest",
      new Uint8Array([1, 2, 3]),
      opts,
      { sleep: noSleep }
    );
    expect(seenUrl).toBe("http://x/ingest");
    expect(seenInit?.method).toBe("POST");
    expect((seenInit?.headers as Record<string, string>)["Authorization"]).toBe(
      "Bearer tok"
    );
  });

  test("uses the real setTimeout sleep by default", async () => {
    let calls = 0;
    const fetchFn = (async () => {
      calls++;
      return calls === 1
        ? new Response("", { status: 500 })
        : new Response("ok");
    }) as unknown as typeof fetch;
    const resp = await postWithRetry(
      fetchFn,
      "http://x/ingest",
      new Uint8Array([1]),
      opts,
      {
        attempts: 2,
        delayMs: 1,
      }
    );
    expect(resp.status).toBe(200);
    expect(calls).toBe(2);
  });
});

const env: Env = {
  INGEST_URL: "http://x/ingest",
  INGEST_TOKEN: "tok",
};

function mockMessage(
  over: Partial<ForwardableEmailMessage> = {}
): ForwardableEmailMessage {
  return {
    from: "sender@x",
    to: "deals@y",
    raw: new Response(new Uint8Array([1, 2, 3])).body!,
    rawSize: 3,
    ...over,
  } as ForwardableEmailMessage;
}

describe("handleEmail", () => {
  test("POSTs the raw body to ingest", async () => {
    let fetchCalls = 0;
    let seenBody: unknown;
    const fetchFn = (async (_url: string, init: RequestInit) => {
      fetchCalls++;
      seenBody = init.body;
      return new Response("ok");
    }) as unknown as typeof fetch;

    await handleEmail(mockMessage(), env, fetchFn, { sleep: noSleep });
    expect(fetchCalls).toBe(1);
    expect((seenBody as ArrayBuffer).byteLength).toBe(3);
  });

  test("propagates a POST failure (no archive, so the message bounces)", async () => {
    const failFetch = (async () => {
      throw new Error("ingest down");
    }) as unknown as typeof fetch;
    await expect(
      handleEmail(mockMessage(), env, failFetch, {
        attempts: 1,
        sleep: noSleep,
      })
    ).rejects.toThrow("ingest down");
  });

  test("throws for an oversize message before POSTing", async () => {
    let fetchCalls = 0;
    const fetchFn = (async () => {
      fetchCalls++;
      return new Response("ok");
    }) as unknown as typeof fetch;
    await expect(
      handleEmail(mockMessage({ rawSize: MAX_RAW_BYTES + 1 }), env, fetchFn)
    ).rejects.toThrow(/exceeds/);
    expect(fetchCalls).toBe(0);
  });

  test("propagates a body-read failure before POSTing", async () => {
    let fetchCalls = 0;
    const erroredStream = new ReadableStream<Uint8Array>({
      start(controller) {
        controller.error(new Error("stream boom"));
      },
    });
    const fetchFn = (async () => {
      fetchCalls++;
      return new Response("ok");
    }) as unknown as typeof fetch;
    await expect(
      handleEmail(mockMessage({ raw: erroredStream }), env, fetchFn)
    ).rejects.toThrow();
    expect(fetchCalls).toBe(0);
  });
});
