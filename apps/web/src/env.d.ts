/// <reference types="astro/client" />

// Client-exposed WorkOS config (see .env-example). Vite inlines VITE_-prefixed
// variables into the browser bundle; declaring them here gives
// `import.meta.env.VITE_*` autocomplete and type-checking. The Go server reads
// the same pair as its WORKOS_CLIENT_ID / WORKOS_API_HOSTNAME fallbacks to
// verify the JWTs these produce.
interface ImportMetaEnv {
  readonly VITE_WORKOS_CLIENT_ID: string;
  readonly VITE_WORKOS_API_HOSTNAME: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
