# Go Image Optimizer

[English](project-guide.md) | [Português](../../pt-BR/guides/project-guide.md)

An image optimization application built with Go, with a web interface using Next.js, React, and Tailwind CSS.

Compression, Resize and format conversion run synchronously, returning processed images directly to the browser. Browser-native previews have a server-generated fallback.

> **Current status:** Complete for the current functional scope: compression, resizing, format conversion and browser-compatible preview fallback. No additional features are planned at this time. Maintenance and bug fixes remain possible; this is not a production-readiness claim.

## Overview

Go Image Optimizer is a portfolio project focused on exploring image processing while applying backend engineering concepts with Go.

Instead of designing a complex architecture upfront, the project follows an incremental approach: start with a simple solution, validate the requirements, and introduce architectural changes when there is a concrete reason for them.

## Current Scope

The application allows users to:

- Upload an image through the web interface.
- Send the image to the Go backend.
- Compress the image, configure Resize by pixels/percentage, or convert its format in a modal.
- Preview supported images natively or through a display-only backend fallback.
- Receive the processed image with measured result information.
- Download the result directly from the browser.

The backend detects the real image format from file bytes rather than trusting extensions or browser MIME labels. Compression and Resize retain the source format family. Compression preserves oriented dimensions, Resize changes them explicitly, and conversion preserves oriented dimensions while changing the family. No operation guarantees a smaller file.

This is a focused Go and Next.js portfolio project. Resize covers basic smaller-image needs. A separate thumbnail operation, processing history, databases, Redis, queues, workers, persistent image/object storage and a Paint-like editor are intentionally outside the scope. The preview fallback is an internal display aid, not a separate user operation.

## Compression Formats

| Format      | Output behavior    | Notes                                                                                                                            |
| ----------- | ------------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| JPEG / JPG  | JPEG               | Lossy re-encode at conservative quality. EXIF orientation is applied to pixels before output.                                    |
| PNG         | PNG                | Lossless pixel output using high PNG compression. Transparency is preserved.                                                     |
| WebP        | WebP               | Static and animated WebP are supported. Animated output preserves frame count and timing while re-encoding reconstructed frames. |
| AVIF        | AVIF               | Static AVIF is supported. The implementation uses the AVIF multi-image API, but current automated coverage uses static fixtures. |
| HEIC / HEIF | HEIC / HEIF family | Uses native libheif/HEVC support in the backend container. Unsupported HEIF variants are rejected instead of being faked.        |
| GIF         | GIF                | Static and animated GIF are supported, including frame delays and loop settings.                                                 |
| BMP         | BMP                | Decoded and re-encoded as BMP; size reduction is not guaranteed.                                                                 |
| TIFF        | TIFF               | Re-encoded as TIFF with Deflate compression.                                                                                     |

WebM, SVG, RAW camera formats, videos, and archives are intentionally not supported.

## Tech Stack

### Backend

- Go
- Native libheif runtime libraries for HEIC/HEIF support in Docker

### Frontend

- Next.js
- React
- Tailwind CSS

## Scope completion

Compression, resizing, conversion and preview fallback are implemented. There is no feature roadmap beyond this scope. Compatibility, native dependencies and resource/cancellation limits remain documented technical limitations. Any hypothetical production evolution would require new requirements and measurements, not extra infrastructure for portfolio complexity.

## Format conversion

Choose a supported static file, inspect its byte-detected family, and select a different destination. Conversion preserves oriented dimensions, with format-specific alpha/metadata policies and measured output sizes; a larger result remains downloadable. Animations and unsupported multi-image variants are rejected. The [architecture](../architecture/overview.md#format-conversion) owns the field/header contract, encoder defaults and variant limits.

## Image previews

Native browser display is attempted first. On failure, a display-only static thumbnail from the backend can help without changing source/download bytes. Preview errors and retry remain local to the preview area. The session cache shares Blob-identity requests; reloading loses transient state. See [preview architecture](../architecture/overview.md#image-previews) for supported variants, resource limits and lifecycle details.

[Testing](../testing/strategy.md) · [Docker](../docker/runtime.md) · [Development setup and recovery](../docker/development.md) · [Verification](../testing/verification.md)
