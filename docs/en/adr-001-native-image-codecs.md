# ADR 001: Native Image Codecs

## Status

Accepted.

## Context

The application now supports HEIC/HEIF in addition to formats that can be handled with Go standard-library or pure-Go codecs.

HEIC/HEIF support is not realistically available in the Go standard library. The selected implementation uses the Go binding for libheif. That gives real HEIC/HEIF decode and encode behavior, but it requires CGO plus native libheif libraries and HEVC plugins at build and runtime.

## Decision

Use native libheif for HEIC/HEIF support and keep the rest of the pipeline synchronous and request-scoped.

The backend Docker image uses:

- Alpine build stage with `CGO_ENABLED=1`;
- `libheif-dev`, `libheif-libde265`, and `libheif-x265` during build;
- Alpine runtime with `libheif`, `libheif-libde265`, and `libheif-x265`;
- a Compose `backend-test` service based on the build stage for local tests that need Go, CGO, and native codec headers;
- no durable image storage, queue, database, or object store.

## Consequences

Positive:

- HEIC/HEIF support is real, not a filename or MIME-type claim.
- Unsupported HEIF variants can be rejected explicitly.
- The rest of the application architecture remains simple and synchronous.

Trade-offs:

- The backend can no longer use a fully static distroless runtime.
- Local native builds need equivalent libheif development libraries installed.
- Runtime images must include the required libheif codec plugins.
- Docker-based backend tests should run in the build-stage test service so real HEIC/HEIF fixtures and synthetic HEIC generation use the same native dependency chain.
- The libheif Go binding writes encoded output through a temporary file API, so the compressor creates and immediately removes an OS temporary file for HEIC/HEIF output.

The decision should be revisited if future requirements demand a fully static runtime, broader HEIF variant support, GPU/native service offload, or asynchronous high-throughput processing.
