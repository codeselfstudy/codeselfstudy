// apiFetch calls the Go JSON API with the caller's WorkOS access token attached
// as a Bearer credential. The server's /api/* routes are gated by the WorkOS
// JWKS middleware, so authenticated requests must carry the token AuthKit issues.
//
// getToken is AuthKit's getAccessToken (from useAuth). When it yields no token
// the request still goes out, just unauthenticated — the server answers 401 —
// so callers can surface that state without a crash rather than throwing here.
export type GetToken = () => Promise<string | undefined | null>;

export async function apiFetch(
  path: string,
  getToken: GetToken,
  init: RequestInit = {}
): Promise<Response> {
  const token = await getToken();
  const headers = new Headers(init.headers);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  return fetch(path, { ...init, headers });
}
