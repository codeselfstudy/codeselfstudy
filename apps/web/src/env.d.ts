/// <reference types="astro/client" />

// Client-exposed WorkOS config (see .env-example). Astro inlines PUBLIC_-prefixed
// variables into the browser bundle; declaring them here gives
// `import.meta.env.PUBLIC_*` autocomplete and type-checking. The Go server reads
// the same pair as its WORKOS_CLIENT_ID / WORKOS_API_HOSTNAME fallbacks to
// verify the JWTs these produce.
interface ImportMetaEnv {
  readonly PUBLIC_WORKOS_CLIENT_ID: string;
  readonly PUBLIC_WORKOS_API_HOSTNAME: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
