# Deployment, diagnostics and evidence

[English](deployment-and-diagnosis.md) | [Português brasileiro](../../pt-BR/operations/deployment-and-diagnosis.md)

Static source review: 2026-10-10. Implemented facts, general theory and hypothetical changes are distinguished below. Runtime commands were not executed.

Compose describes Go backend and Next.js frontend, plus test/development profiles. The backend runtime requires the native codec libraries supplied by its Dockerfile. Development bind mounts and dependency caches support editing; they are not a production rollout procedure. The frontend's server-side `BACKEND_URL` must resolve from its own runtime, typically `backend:8080` inside Compose. `localhost` there means that container, not the developer host or another service.

Diagnose separately: a proxy 502 suggests forwarding/abort/connectivity failure; a backend 4xx identifies invalid or unsupported input; a codec 500 suggests deployment or encoder failure; `/health` is liveness, not an all-format codec self-test. Inspect structured `slog` warnings/errors without logging image contents. Preserve a minimal non-sensitive reproducer and its actual format/variant. Backend HTTP tests, application fakes and imaging fixtures establish different boundaries; recorded historical builds/test results are not new validation. No browser, load, race or application execution occurred in this review.

There is no persistent upload database to back up and no migration process. Updating the backend must preserve matching runtime codec dependencies, and frontend/backend response metadata must remain compatible. Rolling back application images is conceptually possible with retained artifacts, but no release registry, production pipeline, TLS topology, rollback rehearsal or production SLO is evidenced. Benchmark categories are not applicable without measurements: pixel limits and byte-size UI output are not throughput benchmarks. Remaining gaps include concurrent memory peaks, interruption of native work, visual fidelity across representative variants and browser compatibility. Existing test documentation retains the historical native race/checkptr limitation.
