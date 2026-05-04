import { useCallback, useState } from "react";
import { useAuth } from "@workos-inc/authkit-react";

import { useApiFetch } from "@/lib/api";

/**
 * Dev-only TanStack Devtools panel that shows the current WorkOS auth
 * state, lets you copy the access token to the clipboard, and pings
 * /api/me to verify the Go-side validator round-trips. Mounted from
 * `apps/web/src/routes/__root.tsx`. Only loads when TanStackDevtools
 * itself is rendered, which the existing config already gates to dev.
 */
function AuthDevPanel() {
  const { user, isLoading, getAccessToken } = useAuth();
  const apiFetch = useApiFetch();

  const [tokenStatus, setTokenStatus] = useState<string>("");
  const [meStatus, setMeStatus] = useState<string>("");
  const [meBody, setMeBody] = useState<string>("");

  const copyToken = useCallback(async () => {
    const token = await getAccessToken();
    if (!token) {
      setTokenStatus("no token (not signed in or session expired)");
      return;
    }
    try {
      await navigator.clipboard.writeText(token);
      setTokenStatus(`copied (${token.length} chars)`);
    } catch (err) {
      setTokenStatus(`clipboard write failed: ${(err as Error).message}`);
    }
  }, [getAccessToken]);

  const probeMe = useCallback(async () => {
    setMeStatus("…");
    setMeBody("");
    try {
      const res = await apiFetch("/api/me");
      setMeStatus(`${res.status} ${res.statusText}`);
      const text = await res.text();
      setMeBody(text);
    } catch (err) {
      setMeStatus(`fetch threw: ${(err as Error).message}`);
    }
  }, [apiFetch]);

  return (
    <div style={{ padding: 12, fontFamily: "monospace", fontSize: 12 }}>
      <h3 style={{ marginTop: 0 }}>WorkOS auth (dev)</h3>

      <section style={{ marginBottom: 12 }}>
        <strong>state:</strong>{" "}
        {isLoading
          ? "loading"
          : user
            ? `signed in as ${user.email}`
            : "signed out"}
      </section>

      <section style={{ marginBottom: 12 }}>
        <button onClick={copyToken} type="button">
          Copy access token
        </button>
        {tokenStatus && (
          <span style={{ marginLeft: 8, color: "#555" }}>{tokenStatus}</span>
        )}
      </section>

      <section>
        <button onClick={probeMe} type="button">
          GET /api/me
        </button>
        {meStatus && (
          <div style={{ marginTop: 6 }}>
            <strong>{meStatus}</strong>
            {meBody && (
              <pre
                style={{
                  marginTop: 4,
                  padding: 6,
                  background: "#f4f4f4",
                  whiteSpace: "pre-wrap",
                  wordBreak: "break-all",
                }}
              >
                {meBody}
              </pre>
            )}
          </div>
        )}
      </section>
    </div>
  );
}

export default {
  name: "Auth (dev)",
  render: <AuthDevPanel />,
};
