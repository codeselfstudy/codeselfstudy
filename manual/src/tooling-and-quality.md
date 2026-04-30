# Tooling & Quality

## Scripts

```bash
bun run test            # Vitest
bun run test:watch      # Vitest in watch mode
bun run test:coverage   # Vitest with coverage report
bun run lint            # ESLint + Stylelint
bun run format          # Prettier
bun run check           # Prettier + ESLint
```

## Linting and Formatting

- ESLint uses `@tanstack/eslint-config`.
- Stylelint checks CSS and Tailwind usage.
- Prettier is the standard formatter.
- Styling is done with Tailwind CSS.

## Tests and Coverage

Tests run with [Vitest](https://vitest.dev/). `bun run test:coverage`
produces an HTML report under `coverage/`.

## Continuous Integration

`.github/workflows/test.yml` runs `bun run test:coverage` on every pull
request and on pushes to `main`. The coverage HTML report is uploaded as a
build artifact (`coverage-html`) and retained for 14 days.

## Git Hooks

`lefthook.yml` runs formatting on staged files and validates lint/test on push.
If hooks block you, fix the reported files and re-run the command.
