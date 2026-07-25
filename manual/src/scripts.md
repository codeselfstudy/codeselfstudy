# Scripts

Prefer `just` for top-level developer workflows (dev, build, clean, deploy,
database tasks). It wraps the underlying `bun run` and `go run` commands so
you don't have to remember which tool runs what.

## Just

Run Just scripts with [Just](https://just.systems/).

```bash
just <script-name>
```

## Available scripts

List available scripts:

```bash
just
```

Common recipes: `just dev`, `just build`, `just clean`, `just deploy`,
`just ssh`, `just db_migrate`, `just db_status`, `just db_create <name>`.

## Bun

Run `package.json` scripts with [Bun](https://bun.com/).

```bash
bun run <script-name>
```

Bun can also run TypeScript scripts directly without a build step.

```bash
bun run file.ts
```

or render markdown in the terminal for easier reading:

```bash
bun run README.md
```

To install autocompletion functionality, run:

```bash
bun completions
```
