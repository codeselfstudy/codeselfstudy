# Tooling & Quality

## Scripts

```bash
bun run test     # Vitest
bun run lint     # ESLint + Stylelint
bun run format   # Prettier
bun run check    # Prettier + ESLint
```

## Linting and Formatting

- ESLint uses `@tanstack/eslint-config`.
- Stylelint checks CSS and Tailwind usage.
- Prettier is the standard formatter.
- Styling is done with Tailwind CSS.

## Git Hooks

`lefthook.yml` runs formatting on staged files and validates lint/test on push.
If hooks block you, fix the reported files and re-run the command.
