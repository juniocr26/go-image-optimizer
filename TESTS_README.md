# Test Documentation

This document describes the current testing strategy for Go Image Optimizer.

## Strategy

The backend has automated Go tests for the application boundary, image compression implementation, and HTTP contract. The frontend currently relies on the production build plus manual workflow validation; no frontend test framework has been added for this MVP.

The tests avoid a universal assertion that every optimized image must be smaller. Some real images are already optimized. Size-reduction assertions are limited to deterministic fixtures created specifically for that purpose.

## Backend Coverage

Automated backend tests cover:

- compress image use case behavior;
- canceled context handling;
- valid JPEG compression;
- valid PNG compression;
- output images can be decoded;
- dimensions are preserved;
- response format is preserved;
- PNG pixel content remains lossless;
- unsupported input is rejected;
- corrupted image content is rejected;
- decoded pixel safety limit behavior;
- missing image field validation;
- malformed multipart requests;
- non-multipart requests;
- 25 MiB request-size protection;
- successful HTTP response headers;
- generated download filenames.

## Frontend Validation

The frontend build validates TypeScript and production compilation.

Manual UI validation should cover:

- drag-and-drop selection;
- file picker selection;
- JPG and PNG previews;
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

Backend tests with Docker:

```bash
docker run --rm -v "$PWD/backend:/src" -w /src golang:1.27.1-alpine go test ./...
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

Then open the frontend, upload a JPG or PNG, compress it, download the result, and confirm the Compose configuration does not mount an application storage volume for image results.

## Current Limitations

- There are no automated browser interaction tests yet.
- There are no visual quality assertions for JPEG output.
- The frontend workflow is manually validated.
- Docker Compose validation is a smoke test, not a load or scalability test.
- No test coverage percentage is claimed.
