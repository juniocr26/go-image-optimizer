# Test Documentation

This document describes the current testing strategy for Go Image Optimizer.

## Strategy

The backend has automated Go tests for the application boundary, image compression implementation, and HTTP contract. The frontend currently relies on the production build plus manual workflow validation; no frontend test framework has been added for this MVP.

The tests avoid a universal assertion that every optimized image must be smaller. Some real images are already optimized. Size-reduction assertions are limited to deterministic fixtures created specifically for that purpose.

## Backend Coverage

Automated backend tests cover:

- compress image use case behavior;
- canceled context handling;
- valid compression for JPEG/JPG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP, and TIFF;
- output images can be decoded;
- dimensions are preserved;
- response format is preserved;
- PNG pixel content remains lossless;
- lossless WebP transparency remains lossless;
- animated GIF frame count, delays, and loop behavior;
- animated WebP frame count and frame durations;
- JPEG EXIF orientation normalization;
- unsupported input is rejected, including camera RAW signatures;
- corrupted image content is rejected for every supported format signature;
- decoded pixel safety limit behavior for every supported format;
- animated canvas-frame pixel safety limits;
- missing image field validation;
- malformed multipart requests;
- non-multipart requests;
- 50 MiB request-size protection;
- successful HTTP response headers;
- generated download filenames;
- extension spoofing, where response type follows detected bytes rather than the uploaded filename.

## Frontend Validation

The frontend build validates TypeScript and production compilation.

Manual UI validation should cover:

- drag-and-drop selection;
- file picker selection;
- previews for browser-renderable formats;
- placeholder rendering for formats the browser cannot preview, such as many HEIC or TIFF files;
- explicit Compress button behavior;
- disabled duplicate submissions while compressing;
- indeterminate loading state;
- result preview;
- original size, optimized size, and reduction display using real bytes;
- honest handling when the optimized result is not smaller;
- download filename convention;
- reset / another image behavior;
- Blob URL cleanup through selection changes, reset, and unmount.

## How to Run

Backend tests with a local Go toolchain:

```bash
cd backend
go test ./...
```

Local HEIC/HEIF tests require native libheif development libraries and HEVC codec plugins. Docker is the recommended path when those are not installed locally.

Backend tests with Docker:

```bash
docker run --rm -v "$PWD/backend:/src" -w /src golang:1.27.1-alpine sh -lc \
  'apk add --no-cache build-base pkgconf libheif-dev libheif-libde265 libheif-x265 >/dev/null && /usr/local/go/bin/go test ./...'
```

Frontend production build:

```bash
cd frontend
npm run build
```

Docker smoke validation:

```bash
docker compose up --build
```

Then open the frontend, upload representative JPEG/JPG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP, and TIFF files, compress them, download the result, and confirm the Compose configuration does not mount an application storage volume for image results.

## Current Limitations

- There are no automated browser interaction tests yet.
- There are no visual quality assertions for JPEG output.
- The frontend workflow is manually validated.
- Docker Compose validation is a smoke test, not a load or scalability test.
- WebM, SVG, RAW, video, and archive formats are intentionally unsupported and are covered as unsupported-input behavior rather than codec tests.
- No test coverage percentage is claimed.
