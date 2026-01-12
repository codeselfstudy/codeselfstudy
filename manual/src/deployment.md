# Deployment

## Docker

The application is containerized using Docker. The `Dockerfile` is optimized for production builds, utilizing multi-stage builds to minimize image size and speed up build times.

### Build Process

1.  **Base Image**: Uses `oven/bun:alpine` as the base image.
2.  **Dependencies**: Separates production dependencies (`prod-deps`) from build dependencies to optimize caching.
3.  **Build**: Compiles the application using `bun run build`.
4.  **Final Image**: Copies only the necessary artifacts (`.output`, `node_modules`, `public`, `package.json`) to the final image.

### Deployment on Fly.io

The `Dockerfile` is specifically configured for deployment on [Fly.io](https://fly.io).

- **Runtime**: Bun
- **Port**: 8080

To deploy:

```bash
just deploy
```
