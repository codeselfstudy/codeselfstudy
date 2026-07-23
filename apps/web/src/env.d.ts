/// <reference types="astro/client" />

// The browser bundle no longer reads any WorkOS config. Auth runs entirely
// server-side: the Go server performs the WorkOS code exchange and sets a
// first-party session cookie, and the client only fetches the cookie-gated
// /api/me (see src/components/auth/SignInButton.tsx). The server-side WorkOS
// secrets (WORKOS_CLIENT_ID, WORKOS_API_HOSTNAME, WORKOS_API_KEY,
// WORKOS_COOKIE_PASSWORD, APP_BASE_URL) live in the Go process's environment,
// never in this bundle.
