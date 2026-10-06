# Documentation validation — 2026-10-06

No Docker containers were running before this project. With `-f docker-compose.yml -f compose.development.yaml`, `--profile dev config --quiet` and `--profile dev up -d backend frontend-dev` passed. There is no database, migration, seeder, background worker or persistent application upload volume. `--profile test run --rm --no-deps -T backend-test go test -count=1 ./...` passed all packages with tests, including native HEIF and read-only real fixtures. The missing development test image was built from the existing target after an initial registry pull fallback.

`--profile dev exec -T frontend-dev node --test tests/resize-options.test.mjs tests/conversion-size.test.mjs tests/preview-cache.test.mjs` passed all 12 tests; Node emitted module-type warnings. `exec -T frontend-dev npx tsc --noEmit` passed. Backend `/health` and frontend `/` returned HTTP 200 inside the frontend container. Host curl could not reach the published port from the restricted execution context; use the container results below. Production build, browser interaction, visual fidelity, race and load checks were not performed.

Services started here were stopped with the same Compose files and `--profile dev stop frontend-dev backend`. Existing volumes, fixtures and application code were preserved. Next.js regenerated `next-env.d.ts`; its original tracked contents were restored after shutdown, and the newly generated untracked TypeScript build-info file was removed. Bilingual paths and local links/anchors were checked.

Container smoke checks posted the existing PNG fixture through `/api/images/resize/info` and `/api/images/preview`: both HTTP 200. Inspection reported 512 × 512, png, image/png, one frame; preview returned 26,968 bytes. Output remained in memory.
