# Scripts

Prefer `just` for top-level developer workflows (dev, build, clean, deploy,
database tasks). It wraps the underlying `bun run` and `drizzle-kit` commands
so you don't have to remember which tool runs what.

## Just

Run Just scripts with [Just](https://just.systems/).

```bash
just <script-name>
```

## Available scripts

```bash
just --list

# or
just -l
```

Common recipes: `just dev`, `just build`, `just clean`, `just deploy`,
`just ssh`, `just db_generate`, `just db_migrate`, `just db_studio`.

## Bun

Run `package.json` scripts with [Bun](https://bun.sh/).

```bash
bun run <script-name>
```

Bun can also run TypeScript scripts directly without a build step.

```bash
bun run file.ts
```

To install autocompletion functionality, run:

```bash
bun completions
```
