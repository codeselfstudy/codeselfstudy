# syntax=docker/dockerfile:1.7
#
# The web app is built locally (see `just deploy`) and apps/web/dist ships in
# the build context. That avoids re-running `bun install` on Fly's remote
# builder, which was ~10min on a cold cache.

# 1. Build the Go binary, copying the prebuilt static site as ./static.
FROM golang:1.26-alpine AS api
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api ./
COPY apps/web/dist ./static
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server .

# 2. Distroless runtime. CA certs + nonroot user + BusyBox shell (debug
# variant) for `fly ssh console`. The shell adds ~2MB on disk and zero
# steady-state RAM. Swap to `:nonroot` to lock the image down later.
FROM gcr.io/distroless/static-debian12:debug-nonroot
WORKDIR /
COPY --from=api /out/server /server
COPY --from=api /src/static /static
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
