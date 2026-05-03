import { useCallback } from "react";
import { useAuth } from "@workos-inc/authkit-react";

/**
 * Returns a fetch-like function that attaches the WorkOS access token to
 * every call. The Go backend's auth middleware validates the token against
 * its JWKS and rejects unsigned/expired ones with 401 — see
 * `apps/api/internal/auth`.
 *
 * Use it for any same-origin call against `/api/*`. For unauthenticated
 * fetches, just use the global `fetch` directly; this hook is only
 * worthwhile when you need the bearer header.
 */
export function useApiFetch() {
  const { getAccessToken } = useAuth();

  return useCallback(
    async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
      const token = await getAccessToken();
      const headers = new Headers(init?.headers);
      if (token) {
        headers.set("Authorization", `Bearer ${token}`);
      }
      if (!headers.has("Accept")) {
        headers.set("Accept", "application/json");
      }
      return fetch(input, { ...init, headers });
    },
    [getAccessToken]
  );
}
