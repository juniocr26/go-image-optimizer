# Docker

The Docker setup runs the same synchronous, no-persistent-storage workflow as the local application.

## Services

- `backend`: Go API on port `8080`.
- `frontend`: Next.js UI on port `3000`.

`docker-compose.yml` does not mount an application storage volume for uploaded or optimized image results. Images are received, processed, returned, and discarded.

## Backend Image

The backend image uses a CGO-enabled Alpine build because HEIC/HEIF support depends on native libheif and HEVC codec plugins. The runtime stage is pinned to Alpine 3.24 to match the libheif plugin package split used by the Go Alpine image.

Build-stage packages:

- `build-base`
- `pkgconf`
- `libheif-dev`
- `libheif-libde265`
- `libheif-x265`

Runtime packages:

- `libheif`
- `libheif-libde265`
- `libheif-x265`

The previous fully static distroless runtime is not suitable for this feature set because libheif loads native libraries/plugins at runtime.

## Validation Commands

```bash
docker compose config
docker compose build
docker compose up
```

Then open the frontend at `http://localhost:3000` and upload representative JPEG/JPG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP, and TIFF samples.

For backend-only tests in Docker:

```bash
docker run --rm -v "$PWD/backend:/src" -w /src golang:1.27.1-alpine sh -lc \
  'apk add --no-cache build-base pkgconf libheif-dev libheif-libde265 libheif-x265 >/dev/null && /usr/local/go/bin/go test ./...'
```

## Troubleshooting

If HEIC/HEIF encoding fails with a codec or plugin error, confirm the runtime image includes `libheif-x265`. Some Linux distributions split libheif into separate decoder and encoder plugin packages.

If a distribution's package split differs from Alpine, install the equivalent libheif runtime, HEIC decoder, and HEVC encoder packages. `libheif-plugins-all` can be useful for diagnostics, but the project Dockerfile keeps the runtime packages narrower.
