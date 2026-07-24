/**
 * A synchronously-readable hint about who is signed in, so the navbar can paint
 * the right control on its first frame instead of flashing Sign In at someone
 * who is already signed in (#367).
 *
 * Two halves, deliberately:
 *
 * - The **flag** is a cookie the Go server writes and clears in lockstep with
 *   the HttpOnly session cookie (see `authHintCookieName` in
 *   apps/api/internal/session/session.go). Because the server owns it, it
 *   cannot outlive the session — signing out elsewhere, or a session expiring,
 *   takes the flag with it.
 * - The **detail** (username, avatar) lives in localStorage, which no server
 *   can invalidate. It is only ever trusted behind the flag, where being one
 *   fetch out of date is harmless.
 *
 * Neither half is authentication. The flag carries no secret and grants
 * nothing; every protected response still depends on the sealed session cookie.
 *
 * The email is deliberately NOT cached. Keeping it off public surfaces is the
 * whole point of the navbar work (#351), and writing it to disk works against
 * that on a shared machine.
 */

export const AUTH_FLAG_COOKIE = "css_auth";
const ACCOUNT_KEY = "auth:account";

export type AccountHint = { username: string; avatar: string };

/**
 * Whether the server says this browser currently has a session.
 *
 * No try/catch: unlike localStorage, `document.cookie` cannot throw, and this
 * only ever runs in the browser (the navbar island is `client:only`).
 */
export function isSignedInHint(): boolean {
  // Match the full name — a cookie called `xcss_auth` is a different cookie.
  return document.cookie
    .split(";")
    .some((part) => part.trim().startsWith(`${AUTH_FLAG_COOKIE}=`));
}

/** The last known username/avatar, or null if there is nothing usable stored. */
export function readAccountHint(): AccountHint | null {
  try {
    const raw = localStorage.getItem(ACCOUNT_KEY);
    if (!raw) return null;
    const parsed: unknown = JSON.parse(raw);
    // Anything can be in storage — another tab, an older release, a person with
    // devtools open. Only a value of the shape we wrote is worth trusting.
    if (!parsed || typeof parsed !== "object") return null;
    const { username, avatar } = parsed as Record<string, unknown>;
    if (typeof username !== "string" || typeof avatar !== "string") return null;
    return { username, avatar };
  } catch {
    /* storage disabled (private mode, blocked cookies), or not JSON */
    return null;
  }
}

export function writeAccountHint(hint: AccountHint): void {
  try {
    localStorage.setItem(ACCOUNT_KEY, JSON.stringify(hint));
  } catch {
    /* storage full or disabled — the hint is an optimization, not a need */
  }
}

export function clearAccountHint(): void {
  try {
    localStorage.removeItem(ACCOUNT_KEY);
  } catch {
    /* ignore */
  }
}
