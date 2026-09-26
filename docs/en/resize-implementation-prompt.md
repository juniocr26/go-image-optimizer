# Cycle 2 — Image Resize implementation prompt

Implement Image Resize in the current Go Image Optimizer checkout. Inspect current code and follow its feature-based organization; preserve the existing Compression behavior and Home layout.

## User experience
- Add Resize beside Compress in the action selector. Run opens a responsive configuration modal, without processing a resize yet.
- Match the existing navy/green visual language and result modal. Show source preview or the existing fallback, filename, original dimensions, and live target dimensions.
- Provide Pixels and Percentage modes. Pixels starts at the oriented source dimensions, with Keep aspect ratio enabled and Don't enlarge enabled. Editing either dimension while locked calculates the other. Unlocking allows stretching, never cropping. Percentage offers 25%, 50%, and 75% smaller (linear dimensions), defaulting to 50% smaller.
- Obtain authoritative dimensions from a Go inspection endpoint, including formats the browser cannot preview. Show loading, errors, and retry inside the modal; never silently guess dimensions.
- Provide explicit Cancel/close and Resize image actions, keyboard focus containment/restoration, scroll locking, and responsive scrolling. Prevent duplicate submissions. Preserve settings on processing errors. Avoid stacking configuration and result modals.
- Show a result modal with actual original/output dimensions, file sizes, preview/fallback, and download. Preserve the source for subsequent actions and release Blob URLs.

## Backend and API
- Keep synchronous multipart request/response and ephemeral processing. Add POST /images/resize/info for metadata and POST /images/resize for processing, with corresponding Next.js proxy routes.
- Organize resize application logic, HTTP handling, and imaging implementation by feature. Extract only genuinely shared image contracts/native codec helpers and transport utilities as needed.
- Support the existing image families (JPEG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP, TIFF), based on detected bytes. Normalize JPEG EXIF orientation before calculating dimensions. Preserve supported animation timing/loops and alpha; reject unsupported variants explicitly rather than flattening them.
- Validate all options server-side, including finite positive integer dimensions, allowed reductions, rounding to nearest pixel (minimum one), proportion behavior, and no-upscale policy. Enforce existing 50 MiB request, 32M decoded/output pixel, and 64M animation canvas-frame limits. Do not trust client metadata.
- Preserve the source format family and use a safe _resized filename. Same-size requests should return the original bytes rather than re-encode unnecessarily. No promise of smaller file size or improved detail when enlarging.
- Do not add persistence, queues, workers, new frontend test frameworks, or speculative architecture.

## Validation and documentation
- Add meaningful application/HTTP/imaging tests for dimensions, aspect ratio, percentage rounding, invalid/oversized requests, orientation, alpha, animations, safe filenames, and existing real fixtures.
- Run the backend regression suite, TypeScript check and frontend production build; exercise desktop/mobile UI, error recovery, keyboard navigation, and the compression flow.
- Update English/Portuguese READMEs, architecture, API and testing documentation with actual implemented behavior and limitations. Report checks and any unresolved limitations honestly. Do not commit or publish automatically.
