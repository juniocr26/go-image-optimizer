# Resource and native-code boundaries

[English](resource-boundaries.md) | [Português brasileiro](../../pt-BR/security/resource-boundaries.md)

Static source review: 2026-10-10. Implemented facts, general theory and hypothetical changes are distinguished below. Runtime commands were not executed.

Encoded bytes are not decoded memory. The imaging layer bounds static dimensions at 32,000,000 pixels and animated canvas work at 64,000,000 frame-pixels. Dimension validation uses division to avoid multiplication overflow. An RGBA buffer at the static limit alone is about 128 million bytes; originals, destination buffers, frames, proxy buffers and codec allocations add more. Limits apply per operation, with variant-specific checks; they are not a global memory quota or concurrency limit. Some native decode work can precede later checks.

`UseCase.Execute` checks context before and after compression. It does not pass a cancellable context to the compressor. Forwarding an abort signal and setting HTTP timeouts therefore do not prove interruption of a synchronous native codec. The Go server sets 5 s header, 2 min read/write and 60 s idle timeouts. HTTP deadlines bound transport behavior; an encoder can continue consuming CPU before returning. A semaphore or separate worker process could provide admission and isolation, at complexity and communication cost; neither exists.

CGO and native codecs broaden HEIF/AVIF support and add ABI/build/security maintenance. Temporary multipart files are removed after successful parsing; file handles close; HEIF helpers clean up their temporary files. Request-scoped results do not constitute durable storage. Panic recovery in a handler does not sandbox a native library or recover from every native process failure. There is no implemented login, TLS termination or per-user rate limit in the reviewed app. Plan public exposure separately and never claim content validation alone makes hostile uploads safe.
