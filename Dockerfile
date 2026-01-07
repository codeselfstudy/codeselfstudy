# syntax = docker/dockerfile:1

# This is a multi-stage build for deployment on Fly.io
# (not for local development)

# This will crash the deployment if the arg isn't provided.
# Run `just deploy` to deploy to ensure that the bun version is set.
ARG BUN_VERSION=must_be_provided
FROM oven/bun:${BUN_VERSION}-alpine AS base

LABEL fly_launch_runtime="Bun"

WORKDIR /app
ENV NODE_ENV="production"
ENV PORT="8080"

FROM base AS build
RUN apk update && \
  apk add build-base pkgconfig python3 vim
COPY bun.lock package.json ./
RUN bun install
COPY . .
RUN bun --bun run build

# Remove development dependencies
RUN rm -rf node_modules && \
  bun install --ci

FROM base
COPY --from=build /app /app
EXPOSE 8080

CMD [ "bun", "run", "start" ]
