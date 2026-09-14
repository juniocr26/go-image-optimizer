# Docker

The Docker setup runs the same synchronous, no-persistent-storage workflow as the local application.

## Services

- `backend`: Go API on port `8080`.
- `frontend`: Next.js UI on port `3000`.
- `backend-test`: profile-gated backend test runner that targets the Go build stage.

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

The production `backend` service uses the final runtime stage and does not include the Go toolchain or mounted source code. Backend tests run through the separate `backend-test` service, which targets the build stage, keeps CGO and native codec dependencies available, and mounts `./backend` at `/src`.

## Validation Commands

```bash
docker compose config
docker compose build backend
docker compose build backend-test
docker compose up
```

Then open the frontend at `http://localhost:3000` and upload representative JPEG/JPG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP, and TIFF samples.

For backend-only tests in Docker:

```bash
docker compose run --build --rm backend-test
```

The test service is behind the `test` profile, so it is not started by normal `docker compose up`. To pass custom Go test flags through the same native-codec environment:

```bash
docker compose run --build --rm backend-test go test -v ./internal/infrastructure/imaging -run TestCompressorCompressesSupportedStaticFormats
```

`go test -race ./...` is currently blocked by a Go `checkptr` failure inside `github.com/strukturag/libheif` while HEIC/HEIF fixtures are encoded. The normal non-race suite is the supported backend test workflow.

## Troubleshooting

If HEIC/HEIF encoding fails with a codec or plugin error, confirm the runtime image includes `libheif-x265`. Some Linux distributions split libheif into separate decoder and encoder plugin packages.

If a distribution's package split differs from Alpine, install the equivalent libheif runtime, HEIC decoder, and HEVC encoder packages. `libheif-plugins-all` can be useful for diagnostics, but the project Dockerfile keeps the runtime packages narrower.
