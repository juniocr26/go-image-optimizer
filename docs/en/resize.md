# Image Resize — Cycle 2

## User flow

Choose one image, select **Resize**, then click **Run**. A configuration modal opens without changing the Home layout. It reads dimensions from the backend, displays the original preview (or the existing browser fallback), and shows the proposed output dimensions.

- **Pixels** starts at the original, display-oriented dimensions. **Keep aspect ratio** and **Don't enlarge** are enabled initially. Editing either dimension anchors the ratio to that dimension. Unlocking the ratio permits stretching; there is no crop or padding.
- **Percentage** offers **25% smaller**, **50% smaller**, and **75% smaller**, with 50% selected initially. These percentages reduce each linear dimension, not file bytes or total pixel area.
- Dimensions are rounded to the nearest integer, with a minimum of one pixel. A 899 × 1599 image reduced by 50% becomes 450 × 800.
- With the ratio locked, Don't enlarge caps the scale at 1. With it unlocked, each dimension is capped independently at the original value. The output summary always shows the effective dimensions.
- Increasing dimensions is allowed when Don't enlarge is unchecked, within the resource limits. It does not recover detail.
- **Resize image** submits the operation; **Cancel**, the close button, or Escape dismisses configuration while idle. Backdrop clicks do not dismiss it. During processing, settings and dismissal are disabled; loading is indeterminate.
- The result modal shows measured original/output dimensions and file sizes, preview/fallback, and download. It replaces the configuration modal. Its existing explicit-close behavior remains unchanged.
- The selected original stays available for another resize or compression. Configuration reopens at the original defaults. Failed processing keeps the options and shows an inline error; failed inspection has a retry action.

The configuration uses a native modal dialog for focus containment and background inertness, initial heading focus, keyboard-operable tabs, body scroll locking, and focus restoration. On narrow screens it becomes a vertically scrolling single-column dialog with a fixed footer inside the dialog.

## API

Both routes accept `multipart/form-data` with exactly one file field named `image`. Format and dimensions are determined from actual bytes; browser MIME, extension, and client dimension hints are not authoritative.

### POST /images/resize/info

Returns JSON, for example:

```json
{"width":899,"height":1599,"format":"jpeg","contentType":"image/jpeg","frameCount":1}
```

The inspection path validates and decodes the source, without re-encoding or storing it. JPEG EXIF orientation and supported native orientation are reflected in the returned dimensions. This works even when the browser cannot preview the format. Inspection can be relatively expensive for large/native images; the modal shows a loading state. Closing the modal aborts the browser request.

### POST /images/resize

| Field | Contract |
| --- | --- |
| `mode` | Required: `pixels` or `percentage` |
| `width`, `height` | Required in pixels mode: positive integers, each at most 32,000,000; the effective output must also satisfy total pixel limits |
| `axis` | `width` (default) or `height`; last edited dimension used to calculate the locked ratio |
| `keepAspectRatio` | `true` (default) or `false`; applies in pixels mode |
| `withoutEnlargement` | `true` (default) or `false` |
| `reduction` | Required in percentage mode: `25`, `50`, or `75` |

Booleans use literal `true`/`false`; explicitly empty values are rejected. Defaults apply when omitted. Duplicate option fields, extra files, and mixed file/text `image` fields are rejected. The UI serializes valid number inputs as decimal integers (including values entered using exponent notation). Percentage mode ignores width/height and always preserves proportions. The server recalculates targets from the uploaded source; a client cannot override the source dimensions. Resizing is always from the original input, not a previous result.

Success returns image bytes with:

- `Content-Type`, `Content-Length`, and an attachment `Content-Disposition` using a sanitized `*_resized` name and a source-family extension;
- `X-Original-Width`, `X-Original-Height`, `X-Image-Width`, `X-Image-Height` (display-oriented pixels);
- `Cache-Control: no-store`.

Errors use JSON `{ "error": "..." }`: 400 for malformed input/options, 413 for request/pixel/frame budgets, 415 for unsupported formats, 422 for unsupported variants, and 500 for codec/internal failures. The Next.js proxy returns 502 when the backend cannot be reached.

Same-origin Next.js routes are `/api/images/resize/info` and `/api/images/resize`. They forward the multipart body, response status/type, filename, and dimension headers through a shared image-request helper also used by compression.

## Architecture and lifetime

```text
ImageUploadForm → ImageResizeModal
  → Next.js image route → resize_image HTTP handler
    → application/imageresize UseCase
      → imaging/resize Decoder → request-local Source
        → target dimensions → resampling → source-family encoder
```

The resize application defines options, dimension rules, the decoder/source boundary, inspection, and execution. Imaging owns decoded pixels and codec details. Shared image format/result/error types live in `application/imageprocessing`; compression keeps aliases for compatibility. HEIF decode/encode and safe filename handling are shared where both features actually need them.

Inspection and processing are separate synchronous requests. The file is uploaded and decoded again for processing. This deliberately avoids introducing an ID, persistent upload cache, database, Redis, queue, worker, or object storage just to configure one image. Temporary multipart/native-codec files are cleaned up. The browser releases Blob URLs on replacement/reset/unmount.

The application checks request context before/after decoding and processing and between animation frames. Individual codec calls are not forcibly interrupted by context cancellation. This is not background processing or a complete CPU/memory isolation mechanism.

## Image behavior and limits

- JPEG, PNG, WebP, static AVIF, supported single-image HEIC/HEIF, GIF, BMP, and single-page TIFF retain their format family. Existing compression support remains unchanged.
- Resize normalizes JPEG EXIF orientation before computing targets. Native AVIF/HEIF orientation follows the installed codec behavior.
- Catmull–Rom resampling works in premultiplied RGBA. PNG/WebP alpha is preserved; resampling changes pixels and is not a pixel-identical operation. BMP has the limitations of its encoder.
- GIF partial frames are composited using source disposal before resampling, then encoded as full-canvas frames with a web-safe palette, binary transparency, original delays, and original loop count. Palette/color quantization and internal disposal representation can change.
- Animated WebP preserves reconstructed full frames, timing, loop count, background, and ICC when present. EXIF/XMP are not copied by the Resize path, to avoid retaining stale dimensions/orientation. Metadata preservation is not universal.
- **Animated AVIF is rejected.** The installed `gen2brain/avif` v0.6.0 decoder returns frames and delays but does not populate `LoopCount`, so finite-loop preservation cannot be promised. APNG, multi-page TIFF, and unsupported multi-image HEIF are also rejected rather than flattened.
- The full multipart request is capped at **50 MiB**; multipart parsing uses an **8 MiB** memory threshold. Input and output are capped at **32 million pixels**. GIF/WebP animations are capped at **64 million canvas-frame pixels**, including the target size. GIF descriptors are counted before full frame decoding; WebP features/container frame counts are checked before frame decoding.
- If effective dimensions equal the original display dimensions, the original encoded bytes are returned unchanged, retaining their metadata and avoiding unnecessary lossy encoding.
- Otherwise, current encoder settings match the existing compression defaults (JPEG/WebP 82, AVIF/HEIF 60; PNG best compression; TIFF Deflate). A resized output may be larger in bytes. There is no batch, crop, format conversion, history, progress percentage, or promise of detail enhancement.

See [Test documentation](../../TESTS_README.md) for automated coverage and UI validation.
