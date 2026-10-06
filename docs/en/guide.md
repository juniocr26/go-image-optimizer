# Go Image Optimizer

[English](guide.md) | [Português](../pt-BR/guide.md)

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

Select an image, choose **Convert format**, and click **Run** to configure conversion. Byte-based inspection reports the actual source family and oriented dimensions. **Convert image** preserves dimensions and leaves the original selected. Results show source/output formats, measured sizes, reduction, increase or no size change, a preview fallback, and a download of the converted bytes even when larger.

Verified outputs: JPEG (JPG), PNG, WebP, AVIF, HEIC/HEIF (HEVC), GIF, BMP and TIFF. Aliases are not separate algorithms. The source family is disabled and rejected server-side. All animated conversions are rejected; no animation pair is offered. APNG, AVIF sequences and multi-image TIFF/HEIF are also rejected by the shared inspection rules.

JPEG, BMP and HEIC composite transparency onto white. PNG, WebP, AVIF and TIFF preserve alpha; GIF uses the WebSafe palette and binary transparency (50% threshold), which can lose colors and partial alpha. Defaults: JPEG/WebP quality 82, WebP method 4/alpha 100, AVIF quality 60/alpha 100/speed 6, HEVC quality 60, PNG best compression, TIFF Deflate with predictor. Metadata and color profiles are not universally preserved. Orientation follows the shared decoder: normalized JPEG EXIF, AVIF autorotation and codec-applied HEIF transformations; other static formats apply EXIF orientation when recognized by the installed metadata reader. No new native codecs are introduced.

`POST /images/convert` (same-origin proxy `POST /api/images/convert`) accepts exactly one multipart file `image` and one `targetFormat`: `jpeg`, `png`, `webp`, `avif`, `heif`, `gif`, `bmp` or `tiff`. Extra fields, duplicates and same-family targets return 400. Unsupported variants return 422. `POST /images/convert/info` (proxy `/api/images/convert/info`) accepts only `image` and returns `width`, `height`, `format`, `contentType`, and `frameCount`.

Success returns encoded bytes, correct Content-Type/Length, sanitized Content-Disposition with `_converted` and destination extension, Cache-Control no-store, and X-Source-Format, X-Output-Format, X-Original-Width/Height and X-Image-Width/Height headers. Limits: 50 MiB request body, 32 million pixels, 64 million cumulative animation canvas pixels during inspection, and 50 MiB output. Output size is checked after encoding; encoder transient memory is codec-dependent. Processing remains synchronous and stateless, including cleanup of temporary HEIF files.

## Image previews

Uploads, Resize/Convert settings and all result dialogs first try native browser image loading. Display-only MIME correction handles mislabeled JPEG, PNG, GIF, WebP and AVIF bytes without changing the source. If loading fails, the shared component posts the actual File or processed result Blob to the same-origin `/api/images/preview` proxy and Go `POST /images/preview` endpoint (exactly one multipart file named `image`).

The stateless endpoint reuses processing detection, decoding, orientation and resource/variant checks. It returns an aspect-ratio-preserving thumbnail bounded by 1200 × 1200, without cropping or enlargement: PNG for transparency, JPEG for opaque pixels, with the corresponding Content-Type and `Cache-Control: no-store`. Supported static JPEG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP and single-page TIFF inputs can use the fallback. Animations, APNG, AVIF sequences, multipage TIFF and unsupported multi-image HEIF remain explicit unsupported previews; RAW remains unsupported. Native browser previews may still display variants that the fallback rejects.

Loading, errors and **Retry preview** remain within the preview area. Failures do not block operations or downloads. Generated bytes are cached by source Blob identity for the session, concurrent requests are shared, obsolete requests are aborted when their last consumer leaves, stale responses are ignored and display Object URLs are revoked. The cache keeps bytes, not long-lived Object URLs. Original upload bytes and actual processed results remain separate from display-only bytes; downloads, names, formats, sizes and reported dimensions describe the actual files, never the thumbnail. Reloading clears the cache.
