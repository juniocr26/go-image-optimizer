# Docker

The Docker setup runs the same synchronous, no-persistent-storage workflow as the local application.

## Services

- `backend`: Go API on port `8080`.
- `frontend`: Next.js UI on port `3000`.
- `backend-test`: profile-gated backend test runner that targets the Go build stage and runs `go test -v ./...`.

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

The production `backend` service uses the final runtime stage and does not include the Go toolchain or mounted source code. Backend tests run through the separate `backend-test` service, which targets the build stage, keeps CGO and native codec dependencies available, mounts `./backend` at `/src`, and mounts `./storage/testdata/images` read-only at `/testdata/images`.

The `storage/testdata/images` mount contains versioned real image fixtures used by tests. It is not application upload storage, and compressed test outputs are not written there.

## Build and Test Workflow

Build the application images when setting up the project or after changing Dockerfiles, native packages, Go dependencies, or frontend dependencies:

```bash
docker compose build
```

For backend-only test setup, building the test image is enough:

```bash
docker compose build backend-test
```

Normal backend test runs reuse that image and do not need `--build`:

```bash
docker compose run --rm backend-test
```

Use `docker compose run` for `backend-test` because the test container is short-lived and profile-gated. Use `docker compose exec backend ...` only for commands that must run inside an already-running backend API container; that runtime container intentionally has no Go toolchain.

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
docker compose run --rm backend-test
```

The test service is behind the `test` profile, so it is not started by normal `docker compose up`. To pass custom Go test flags through the same native-codec environment:

```bash
docker compose run --rm backend-test go test -v ./internal/infrastructure/imaging -run TestCompressorRealFixtures
```

`go test -race ./...` is currently blocked by a Go `checkptr` failure inside `github.com/strukturag/libheif` while HEIC/HEIF fixtures are encoded. The normal non-race suite is the supported backend test workflow.

## Troubleshooting

If HEIC/HEIF encoding fails with a codec or plugin error, confirm the runtime image includes `libheif-x265`. Some Linux distributions split libheif into separate decoder and encoder plugin packages.

If a distribution's package split differs from Alpine, install the equivalent libheif runtime, HEIC decoder, and HEVC encoder packages. `libheif-plugins-all` can be useful for diagnostics, but the project Dockerfile keeps the runtime packages narrower.
