# Archived: meetup feature issues (v1)

**Archived:** 2026-07-22
**Status:** all closed / superseded

These are the 18 GitHub issues that were labeled `meetup` (#118–#136), captured verbatim before being closed. They described a first-party Meetup.com replacement (events + RSVPs + AI recaps) planned against the site's **previous** architecture:

> TanStack Start + oRPC + Drizzle/Turso (libsql) + WorkOS AuthKit + shadcn/ui + Gemini Flash

That stack has since been retired. The site now runs **Astro SSG + Go Echo** (landed on `main` as v4.1.0 via #233), so every implementation detail below — file paths (`src/routes/*.tsx`, `src/server/*`, Drizzle schema), oRPC routers, `createServerFileRoute`, etc. — no longer maps to the codebase. The parent epic (#136) also references a local design doc (`~/.claude/plans/i-ve-brainstormed-this-spec-breezy-hoare.md`) that no longer exists.

These issues are kept here only as a **record of the intended feature scope and design decisions** (waitlist FIFO promotion, RSVP-close semantics, the draft→generate→edit→publish recap flow, the admin/member role model). The _screens_ are stack-agnostic and carry forward; a fresh plan for the Astro + Go implementation, plus screen mock-ups (`mockups/75-meetup-attendee.html`, `mockups/76-meetup-admin.html`), replace this work.

Bodies are reproduced exactly as they appeared on GitHub (including their original backtick escaping and now-stale links).

## Contents

**Parent epic:** #136 — Event management & RSVP system (Meetup.com replacement)

**Foundation:** #118 (deps + Gemini env) · #119 (DB schema) · #120 (WorkOS JWT verify + user upsert) · #121 (oRPC bootstrap) · #122 (client oRPC + token sync) · #123 (rsvp/slug lib helpers) · #124 (seed script)

**Server procedures:** #125 (events router) · #126 (RSVP router + waitlist) · #127 (users router) · #128 (Gemini integration) · #129 (summary router)

**UI:** #130 (public events list) · #131 (public event detail + RSVP) · #132 (admin layout + event form) · #133 (admin recap editor) · #134 (admin users)

**Docs:** #135 (manual page)

---

## #118 — Install server deps and add Gemini env vars

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/118

Part of the Meetup replacement project (parent issue to follow).

## Scope

- \`bun add @workos-inc/node jose @google/genai\` — confirm \`@google/genai\` package name on npm before installing.
- Update [src/env.ts](src/env.ts) to add:
  - \`GEMINI_API_KEY: z.string()\` (server)
  - \`GEMINI_MODEL: z.string().default("gemini-2.5-flash")\` (server, optional with default)

## Done when

- \`bun run build\` succeeds.
- \`bun run dev\` boots with the new vars set in \`.env\`.

---

## #119 — DB schema: groups, users, events, rsvps, event_summaries

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/119

Part of the Meetup replacement project.

## Scope

Append the following Drizzle tables to [src/db/schema.ts](src/db/schema.ts) (keep existing \`todos\` untouched):

- **groups**: id, slug unique, name, description?, createdAt.
- **users**: id, workosId unique, email, firstName?, lastName?, profilePictureUrl?, role (enum \`member|admin\`, default \`member\`), timestamps. Index on email.
- **events**: id, groupId FK, slug, title, description (markdown), location?, isOnline bool, category enum (\`study-hall|workshop|social|talk|other\`), startTime, endTime, capacity (nullable = unlimited), rsvpClosesAt?, rsvpsClosed bool, status enum (\`draft|published|cancelled\`), createdById FK → users, timestamps. Unique \`(groupId, slug)\`. Indexes on startTime and status.
- **rsvps**: id, eventId FK cascade, userId FK cascade, status enum (\`confirmed|waitlisted|cancelled\`), waitlistedAt?, confirmedAt?, cancelledAt?, notes?, timestamps. Unique \`(eventId, userId)\`. Index \`(eventId, status, waitlistedAt)\`.
- **event_summaries**: id, eventId unique FK, rawNotes, aiDraft?, aiDraftGeneratedAt?, aiModel?, publishedSummary?, publishedAt?, publishedById? FK → users, updatedAt.

## Done when

- \`bun run db:generate\` produces a migration with FKs and indexes above.
- Migration applies cleanly to Turso dev via \`bun run db:push\`.
- \`bun run lint\` and \`bun run build\` pass.

---

## #120 — Server context: WorkOS JWT verify + lazy users upsert

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/120

Part of the Meetup replacement project.

## Scope

Create \`src/server/context.ts\`:

- Parse \`Authorization: Bearer <jwt>\` from incoming request.
- Verify against WorkOS JWKS via \`jose.createRemoteJWKSet\` at \`https://{VITE_WORKOS_API_HOSTNAME}/sso/jwks/{VITE_WORKOS_CLIENT_ID}\`.
- On valid token: lazy-upsert the local \`users\` row (match on \`workosId\`; sync email/name/picture when they change).
- Return \`{ db, user: { id, workosId, email, role } | null }\`.

## Done when

- Unit test (or a tiny integration test with an in-memory DB) covers: no token → user null; valid token → row inserted; existing row with changed email → row updated.
- Invalid/expired token returns \`user: null\` without throwing.

Depends on DB schema issue.

---

## #121 — oRPC bootstrap: base procedures + fetch handler route

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/121

Part of the Meetup replacement project.

## Scope

- Create \`src/server/orpc.ts\` exposing:
  - \`publicProcedure\` = \`os.\$context<AppContext>()\`
  - \`authedProcedure\` → throws \`ORPCError("UNAUTHORIZED")\` if no user
  - \`adminProcedure\` → builds on authed, throws \`FORBIDDEN\` if \`role !== "admin"\`
- Create \`src/server/router.ts\` as an empty root router (child routers added in later issues).
- Create \`src/routes/api/rpc.\$.ts\` using \`createServerFileRoute(\"/api/rpc/\$\").methods({ GET, POST, PUT, PATCH, DELETE })\` delegating to \`new RPCHandler(router).handle(request, { prefix: \"/api/rpc\", context })\`.

## Flags

- Verify \`createServerFileRoute\` export path against the pinned \`@tanstack/react-start ^1.145.9\` version.

## Done when

- A smoke \`publicProcedure\` (e.g. \`ping\`) can be hit via \`curl\` at \`/api/rpc/ping\` and returns a valid response.

Depends on server-context issue.

---

## #122 — Client oRPC integration + access-token sync

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/122

Part of the Meetup replacement project.

## Scope

- \`src/integrations/orpc/client.ts\`: \`RPCLink\` with \`headers()\` reading an access-token ref; \`createTanstackQueryUtils(client)\` exports \`orpc\`.
- \`src/integrations/orpc/provider.tsx\`: \`<AccessTokenSync/>\` component that writes \`useAuth().getAccessToken()\` to the module-level ref on change.
- Mount \`<AccessTokenSync/>\` inside \`WorkOSProvider\` in [src/routes/__root.tsx](src/routes/__root.tsx).

## Flags

- Confirm no first-paint race; if fragile, pivot to a \`useORPCClient()\` hook closing over \`useAuth()\`.

## Done when

- A test call to an \`authedProcedure\` from a React component succeeds after sign-in and returns 401 when signed out.

Depends on oRPC bootstrap issue.

---

## #123 — lib helpers: rsvp (close + promoteWaitlist) and slug

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/123

Part of the Meetup replacement project.

## Scope

- \`src/lib/rsvp.ts\`:
  - \`isRsvpEffectivelyClosed(event, now?)\` — returns true when \`rsvpsClosed\` OR \`now >= (rsvpClosesAt ?? startTime)\`.
  - \`promoteWaitlist(tx, eventId, slotsToFill)\` — promotes the N oldest waitlisted RSVPs (ORDER BY \`waitlistedAt\` ASC) to confirmed; clears \`waitlistedAt\`, sets \`confirmedAt\`.
- \`src/lib/slug.ts\`:
  - \`slugify(title)\` — kebab-case, ASCII, collapse dashes, trim length.
  - Caller-side retry pattern (document) that appends \`-2\`, \`-3\`, ... on unique-constraint conflict when inserting events.

## Done when

- \`src/lib/rsvp.test.ts\` covers the effective-closed matrix and exercises promoteWaitlist over an in-memory libsql DB.
- \`src/lib/slug.test.ts\` covers basic cases and collision suffix logic.

Depends on DB schema issue.

---

## #124 — Seed script: Code Self Study group + initial admin

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/124

Part of the Meetup replacement project.

## Scope

\`src/scripts/seed.ts\`, runnable via \`bun run src/scripts/seed.ts\`:

- Upsert \`groups\` row: \`{ slug: \"code-self-study\", name: \"Code Self Study\" }\`.
- Read \`INITIAL_ADMIN_EMAIL\` env and promote the matching \`users\` row to \`role = 'admin'\`.

## Notes

- The admin user must sign in once (via the WorkOS flow) before the script can find their row — that sign-in creates the row via the lazy upsert. Document this in the script's header comment.

## Done when

- Running the script after a fresh DB + first admin sign-in results in a group row and an \`admin\`-role user.

Depends on DB schema + server context issues.

---

## #125 — Events router: CRUD, publish, cancel, list

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/125

Part of the Meetup replacement project.

## Scope

\`src/server/routers/events.ts\` exposing:

- \`listUpcoming\` (public) — \`status = 'published' AND startTime >= now()\`, ordered asc. Returns event + confirmed count + effective-closed flag.
- \`listPast\` (public) — \`endTime < now()\`.
- \`bySlug\` (public) — returns event + counts; if caller is authed, include their RSVP row. Drafts visible to admins only (404 otherwise).
- \`listAdmin\` (admin) — all events incl. drafts + cancelled.
- \`create\` (admin) — full form schema, slug auto-generated (using \`src/lib/slug.ts\` with collision retry).
- \`update\` (admin) — if \`capacity\` rises, call \`promoteWaitlist\` inside the same tx.
- \`publish\` (admin) — sets \`status = 'published'\`.
- \`cancel\` (admin) — sets \`status = 'cancelled'\` and marks all non-cancelled RSVPs as cancelled.

## Done when

- Zod input schemas reject bad inputs (endTime before startTime, negative capacity, etc.).
- Procedures can be called from a React component via the oRPC client.

Depends on oRPC bootstrap + lib helpers.

---

## #126 — RSVP router: create/cancel + waitlist promotion + concurrency tests

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/126

Part of the Meetup replacement project.

## Scope

\`src/server/routers/rsvp.ts\`:

- \`create\` (authed) — wraps in \`db.transaction\`. If \`event.status !== 'published'\` or effective-closed → \`BAD_REQUEST\`. Capacity unlimited or \`confirmedCount < capacity\` → confirmed; else waitlisted with \`waitlistedAt = now\`. Upserts the row by \`(eventId, userId)\`.
- \`cancel\` (authed) — transactionally updates to cancelled; if previous status was confirmed and capacity is finite, promote head-of-waitlist (FIFO via \`waitlistedAt\`).
- \`myRsvps\` (authed) — user's RSVPs with event info.
- \`listForEvent\` (admin) — ordered attendee roster (confirmed + waitlisted in FIFO).

## Tests (\`src/server/routers/rsvp.test.ts\`, in-memory libsql)

- Under capacity → confirmed.
- At capacity → waitlisted.
- Cancel confirmed → head of waitlist auto-promoted.
- Cancel waitlisted → no promotion.
- Two concurrent \`create\` calls at \`capacity - 1\` via \`Promise.all\` → exactly one confirmed, one waitlisted.
- Re-RSVP after cancel reuses the row.
- Effective-closed blocks new RSVPs.
- Capacity raise promotes in FIFO order.

## Flags

- Validate libsql remote-HTTP transaction behavior — if serialization breaks, add advisory-lock table.

Depends on events router + lib helpers.

---

## #127 — Users router: me, list, setRole with last-admin guard

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/127

Part of the Meetup replacement project.

## Scope

\`src/server/routers/users.ts\`:

- \`me\` (authed) — current user row.
- \`list\` (admin) — paginated with optional \`query\` (name/email prefix).
- \`setRole\` (admin) — \`{ userId, role }\`. Refuse when:
  - \`userId\` is the caller (no self-demote), OR
  - \`role = 'member'\` and caller would be leaving the system with zero admins.

## Done when

- Unit tests cover the two guard conditions.
- Procedure is callable from the admin users UI.

Depends on oRPC bootstrap.

---

## #128 — Gemini integration: client + summarize helper

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/128

Part of the Meetup replacement project.

## Scope

- \`src/integrations/gemini/client.ts\` — init \`@google/genai\` with \`env.GEMINI_API_KEY\`.
- \`src/integrations/gemini/summarize.ts\` — \`summarizeEventNotes({ event, rawNotes }) => { text, model }\`.
  - Model from \`env.GEMINI_MODEL\` (default \`gemini-2.5-flash\`).
  - System prompt: factual, 2–4 markdown paragraphs, friendly tone, \"do not invent attendees, projects, or quotes.\"
  - Guardrails: reject \`rawNotes.length < 50\`; cap raw input at 20k chars.
  - Wrap SDK errors in \`ORPCError(\"INTERNAL_SERVER_ERROR\")\`.

## Done when

- Unit test with a mocked SDK asserts prompt shape and error propagation.

## Flags

- Confirm current Gemini SDK package name (\`@google/genai\` vs \`@google/generative-ai\`) and current Flash model ID at install time.

---

## #129 — Summary router: notes, generate, edit, publish, unpublish

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/129

Part of the Meetup replacement project.

## Scope

\`src/server/routers/summary.ts\`:

- \`get\` (public) — returns \`publishedSummary\` only; admins get full row incl. \`rawNotes\` and \`aiDraft\`.
- \`saveNotes\` (admin) — upsert \`rawNotes\`.
- \`generateDraft\` (admin) — call Gemini, write \`aiDraft\`, \`aiDraftGeneratedAt\`, \`aiModel\`; return draft. Dedupe: reject if \`aiDraftGeneratedAt\` is within 60s.
- \`updateDraft\` (admin) — hand-edit the draft text.
- \`publish\` (admin) — sets \`publishedSummary\`, \`publishedAt\`, \`publishedById\`.
- \`unpublish\` (admin) — clears the three above.

## Done when

- End-to-end flow works: save notes → generate → edit → publish → visible via public \`get\`.

Depends on Gemini integration + events router.

---

## #130 — Public /events page: tabbed upcoming/past list (replace static)

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/130

Part of the Meetup replacement project.

## Scope

Replace [src/routes/events.tsx](src/routes/events.tsx):

- Remove the static \`EventsContent\` import.
- Tabs: \"Upcoming\" / \"Past\".
- Card grid (shadcn \`Card\`) showing title, category badge, start time, location, confirmed count / capacity.
- SSR via TanStack Router \`loader\` + \`queryClient.ensureQueryData(orpc.events.listUpcoming.queryOptions())\` (and past).
- Use \`PageWrapper\` + existing shadcn components.

## Done when

- Upcoming and past events render from the DB.
- First paint is hydrated (no client-side loading flash).
- Page still passes \`bun run lint\`.

Depends on client oRPC + events router.

---

## #131 — Public event detail page with RSVP button

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/131

Part of the Meetup replacement project.

## Scope

\`src/routes/events.\$eventSlug.tsx\`:

- Title, description (markdown render), category badge, location, online indicator, start/end with explicit TZ label.
- Confirmed count vs capacity + waitlist size if applicable.
- RSVP button states:
  - Not signed in → \"Sign in to RSVP\" (no redirect).
  - Can RSVP → \"RSVP\".
  - Already confirmed → \"Cancel RSVP\" (with \`AlertDialog\` confirm).
  - On waitlist → shows waitlist position + \"Leave waitlist\".
  - Effective-closed / cancelled → disabled with reason.
- Published summary block (markdown) when present.

## Done when

- All state transitions verified manually in dev.
- Page handles draft events (404) and cancelled events (readonly banner) correctly.

Depends on public /events page + rsvp router.

---

## #132 — Admin layout + events list + create/edit form

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/132

Part of the Meetup replacement project.

## Scope

- \`src/routes/admin.tsx\` — pathless layout with \`beforeLoad\` that calls \`orpc.users.me.queryOptions()\`; \`redirect({ to: \"/\" })\` if \`role !== \"admin\"\`.
- \`src/routes/admin/events.index.tsx\` — admin list incl. drafts/cancelled; each row links to edit + summary.
- \`src/routes/admin/events.new.tsx\` and \`events.\$eventId.edit.tsx\` — shared form component using TanStack Form + shadcn (\`input\`, \`textarea\`, \`select\`, \`calendar\`, \`switch\` for \`isOnline\` and \`rsvpsClosed\`). Fields: title, description, category, startTime, endTime, capacity (empty = unlimited), location, isOnline, rsvpClosesAt, status.

## Done when

- Admin can create, edit, publish, and cancel events end-to-end in dev.
- Non-admins hitting \`/admin/*\` are redirected to \`/\` with no flash.

Depends on events + users routers.

---

## #133 — Admin summary page: draft \u2192 generate \u2192 edit \u2192 publish

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/133

Part of the Meetup replacement project.

## Scope

\`src/routes/admin/events.\$eventId.summary.tsx\`:

- Left panel: raw notes textarea with a \"Save notes\" button.
- Right panel: draft markdown editor.
- Controls: \"Generate draft\" (calls Gemini via \`summary.generateDraft\`), \"Publish\" / \"Unpublish\".
- Shows \`aiDraftGeneratedAt\` + model used.
- Toasts for save / generation success/failure (use \`sonner\`, already imported).

## Done when

- Full flow round-trips in dev: notes save, draft generates, edits persist, publish renders on public page.
- Rejected short notes (<50 chars) surface a clear error toast.

Depends on summary router.

---

## #134 — Admin users page with role toggle

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/134

Part of the Meetup replacement project.

## Scope

\`src/routes/admin/users.tsx\`:

- Table of users (shadcn \`Table\`): name, email, role, joined date.
- Search box filtering by name/email.
- Role dropdown per row calling \`users.setRole\`.
- Self-row and last-admin rows show a disabled dropdown with tooltip explaining why.

## Done when

- Admin can promote/demote other members.
- Guard errors (self-demote, last-admin) surface as toasts.

Depends on users router.

---

## #135 — Documentation: manual page for events & RSVP

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/135

Part of the Meetup replacement project.

## Scope

- Add \`manual/src/events-and-rsvp.md\` covering:
  - High-level architecture (groups, users, events, rsvps, summaries).
  - Admin workflows: create event, manage RSVPs, summarize.
  - Member workflows: sign in, RSVP, waitlist behavior, cancellation.
  - RSVP close semantics (\`rsvpsClosed\`, \`rsvpClosesAt\`, startTime fallback).
  - Waitlist FIFO + auto-promotion rules.
  - A Mermaid sequence diagram for the RSVP-with-waitlist flow.
- Link from \`manual/src/SUMMARY.md\`.

## Done when

- \`just build_manual\` (or equivalent) renders without errors.
- Page is linked in the left-nav of the built manual.

---

## #136 — Event management & RSVP system (Meetup.com replacement)

- **State:** OPEN
- **URL:** https://github.com/codeselfstudy/codeselfstudy/issues/136

## Context

Code Self Study currently relies on Meetup.com for events and RSVPs. This project replaces that with a first-party event + RSVP system inside \`codeselfstudy.com\`, built on the existing stack (TanStack Start, Drizzle + Turso/libsql, WorkOS AuthKit, shadcn/ui). Goal: eliminate the recurring Meetup platform cost and give us an integrated experience for members.

Full design doc: `~/.claude/plans/i-ve-brainstormed-this-spec-breezy-hoare.md` (local planning file).

## Scope (MVP)

- Groups, users (local mirror of WorkOS with member/admin role), events, RSVPs, AI-polished event summaries.
- Waitlist with FIFO auto-promotion on cancellation.
- RSVP close via optional \`rsvpClosesAt\` + manual \`rsvpsClosed\` boolean (falls back to start time).
- Gemini Flash (\`gemini-2.5-flash\`) for draft → edit → publish summary flow.

**Out of scope for MVP** (phase 2): recurring events, email (confirmations, reminders, promotion notices), ICS / calendar exports, cron auto-close, multi-tag join tables, event cover images, check-in tracking.

## Resolved design decisions

- **Admin model**: DB-stored role on local \`users\` table mirroring WorkOS users. Lazy-upsert on first authenticated request.
- **Waitlist**: auto-promote FIFO when a confirmed RSVP cancels or capacity is raised.
- **RSVP close**: optional \`rsvpClosesAt\` + manual \`rsvpsClosed\` toggle. Default fallback is event \`startTime\`.
- **AI summary flow**: Draft → Generate → Edit → Publish. Store raw notes, last AI draft, and final published summary separately.

## Sub-issues

### Foundation

- [ ] #118
- [ ] #119
- [ ] #120
- [ ] #121
- [ ] #122
- [ ] #123
- [ ] #124

### Server procedures

- [ ] #125
- [ ] #126
- [ ] #127
- [ ] #128
- [ ] #129

### UI

- [ ] #130
- [ ] #131
- [ ] #132
- [ ] #133
- [ ] #134

### Docs

- [ ] #135

## Dependency ordering

\`#118 → #119 → #120 → #121 → #122\` is the hard-serial foundation. Then \`#123\` (lib helpers) and \`#124\` (seed) can go in parallel. After that, router issues (\`#125\`–\`#129\`) and UI issues (\`#130\`–\`#134\`) can mostly proceed in parallel within their tiers. \`#135\` (docs) last.

## Open risks (tracked in sub-issues)

- \`createServerFileRoute\` export path against the pinned TanStack Start version.
- AuthKit-react access-token plumbing (ref-based vs hook-based).
- \`@google/genai\` package name + current Flash model ID.
- libsql remote-HTTP transaction semantics under concurrent RSVP writes.

---
