import {
  defineEventHandler,
  getRequestHeader,
  getRequestHeaders,
  proxyRequest,
} from "h3";

const GO_TIMER_SERVER = process.env.TIMER_SERVER_URL || "http://localhost:8081";

// Proxy WebSocket connections to the Go timer server
export default defineEventHandler(async (event) => {
  const path = event.context.params?.path || "";
  const target = `${GO_TIMER_SERVER}/ws/${path}`;

  console.log(
    `[WS Proxy] Request received for: ${event.path} -> Target: ${target}`
  );

  // Check if this is a WebSocket upgrade request
  const isWebSocket =
    getRequestHeader(event, "upgrade")?.toLowerCase() === "websocket";

  if (isWebSocket) {
    console.log(`[WS Proxy] Upgrading connection to: ${target}`);
    // For WebSocket upgrades, we need to proxy using fetch with upgrade
    const headers = getRequestHeaders(event);

    return proxyRequest(event, target, {
      headers: {
        ...headers,
        host: new URL(target).host,
      },
    });
  }

  // For non-WebSocket requests (health checks, etc.), proxy normally
  return proxyRequest(event, target);
});
