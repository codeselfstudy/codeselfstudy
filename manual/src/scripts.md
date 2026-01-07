# Scripts

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
