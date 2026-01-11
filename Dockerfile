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

# Stage 1: Install production dependencies
FROM base AS prod-deps
COPY bun.lock package.json ./
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --ci --production

# Stage 2: Build the application
FROM base AS build
RUN --mount=type=cache,target=/var/cache/apk \
    apk update && apk add build-base pkgconfig python3
COPY bun.lock package.json ./
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --ci
COPY . .
RUN bun --bun run build

# Stage 3: Final image
FROM base
# Copy production dependencies
COPY --from=prod-deps /app/node_modules /app/node_modules
# Copy built application
COPY --from=build /app/.output /app/.output
COPY --from=build /app/public /app/public
COPY --from=build /app/package.json /app/package.json
COPY dotfiles/* /root/

EXPOSE 8080

CMD [ "bun", "run", "start" ]
