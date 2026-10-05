# Test Data

[English](README.md) | [Português](README.pt-BR.md)

`storage/testdata/images` contains versioned real image fixtures for backend integration tests.

These files are test inputs only. They are not application upload storage, compressed outputs, processing history, or user data. The application still processes uploads synchronously and returns processed bytes directly to the caller without persisting them under `storage`.

Compressed test results must stay in memory or in Go test temporary paths such as `t.TempDir()`. Do not write generated outputs back into `storage/testdata/images`.


## Physical sample inventory

All nine versioned processing samples are static 512 × 512 images. Both HEIC/HEIF files use the same format family.

| File | Bytes | Family |
| --- | --- | --- |
| [sample.avif](images/sample.avif) | 4351 | avif |
| [sample.bmp](images/sample.bmp) | 1048714 | bmp |
| [sample.gif](images/sample.gif) | 11596 | gif |
| [sample.heic](images/sample.heic) | 6622 | heif |
| [sample.heif](images/sample.heif) | 8391 | heif |
| [sample.jpg](images/sample.jpg) | 30262 | jpeg |
| [sample.png](images/sample.png) | 24487 | png |
| [sample.tiff](images/sample.tiff) | 1048946 | tiff |
| [sample.webp](images/sample.webp) | 7432 | webp |

Compression and preview already use all nine files. Resize covers four transformations per sample plus no-op byte preservation. Conversion covers nine sources × eight targets, including nine expected same-family errors. Outputs stay in memory, and the Docker mount is read-only.

The four versioned UI assets in `frontend/public/images` (`branding/favicon.ico`, `branding/logo.png`, `hero/grassfield.png`, `hero/mountain.png`) are branding/background assets, not processing test samples; they are also unchanged.
