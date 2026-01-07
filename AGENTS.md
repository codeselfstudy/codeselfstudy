# Repository Guidelines

## Project Structure & Module Organization

This is a Vite + React (TanStack Start) app with TypeScript.

- `src/` holds application code.
- `src/routes/` defines route entries (TanStack Router).
- `src/components/` contains reusable UI components.
- `src/content/` and `src/data/` store site content and data helpers.
- `src/db/` contains database schema and access code (Drizzle).
- `public/` contains static assets.
- `manual/` contains the mdBook manual content. Write documentation for any changed code. Mermaid is available.

## Build, Test, and Development Commands

Use bun. Do not use npm, pnpm, or yarn.lock`.

- `bun run dev`: start the dev server at `http://localhost:3000`.
- `bun run build`: create the production build.
- `bun run preview`: serve the production build locally.
- `bun run test`: run Vitest once (passes if no tests exist).
- `bun run lint`: run ESLint and Stylelint (auto-fix enabled).
- `bun run format`: format code with Prettier.
- `bun run db:generate|db:migrate|db:push|db:pull|db:studio`: Drizzle database tasks.
- `just find_licenses`: list embedded `LICENSE.md` files.

## Coding Style & Naming Conventions

- TypeScript + React, module syntax is ESM.
- Indentation is 2 spaces (follow existing files).
- Components and hooks use `PascalCase`/`camelCase`.
- Routes follow the TanStack conventions in `src/routes/`.
- Formatting and linting are handled by `prettier`, `eslint`, and `stylelint`.

## Testing Guidelines

- Framework: Vitest (`vitest.config.ts`).
- Name tests `*.test.ts` or `*.test.tsx` near the module under test.
- Run `bun run test` before opening a PR; add tests for new behavior when practical.

## Commit & Pull Request Guidelines

- Commit messages in history are short, sentence-case summaries (often with a period). Follow that style.
- PRs should include: a concise summary, testing notes (`bun run test`, `bun run lint`), and screenshots for UI changes.

## Configuration & Environment

- Environment variables are validated in `src/env.ts`.
- Required server vars: `SERVER_URL`, `WORKOS_API_KEY`, and `DATABASE_URL` (for Drizzle).
- Client-side variables must be prefixed with `VITE_`.
- Don't read files or directories ending in `.bak` or that are blocked by `.gitignore`.
