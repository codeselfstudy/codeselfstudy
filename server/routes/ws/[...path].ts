import {
  defineEventHandler,
  getRequestHeaders,
  isWebSocketUpgradeRequest,
  proxyRequest,
} from "h3";

const GO_TIMER_SERVER = process.env.TIMER_SERVER_URL || "http://localhost:8081";

// Proxy WebSocket connections to the Go timer server
export default defineEventHandler(async (event) => {
  const path = event.context.params?.path || "";
  const target = `${GO_TIMER_SERVER}/ws/${path}`;

  // Check if this is a WebSocket upgrade request
  if (isWebSocketUpgradeRequest(event)) {
    // For WebSocket upgrades, we need to proxy using fetch with upgrade
    const headers = getRequestHeaders(event);

    // Note: In production with Bun runtime, WebSocket proxying is handled
    // automatically by proxyRequest when the proper headers are present.
    // For development, we pass through the upgrade headers.
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
