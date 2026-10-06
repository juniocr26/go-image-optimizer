# Docker development setup and recovery

Run all commands from the host `go-image-optimizer/` directory. Require Docker Engine/Desktop, Compose supporting multiple files and profiles (audit: v5.1.4), network access to npm/Go/Alpine registries, and adequate build memory. No host Go or npm is needed. The backend uses Go 1.27.1 Alpine with CGO; the frontend image uses the floating `node:24-alpine` tag. Exact installed frontend versions come from `frontend/package-lock.json`, not the ranges in package.json.

## Services and startup behavior

`backend` serves the Go image-processing API on 8080. `frontend` is the production Next.js standalone server on 3000. `frontend-dev` (profile `dev`) runs Next.js development mode; do not start it alongside `frontend`, since both publish 3000. `backend-test` (profile `test`) normally runs Go tests and mounts image fixtures read-only. It is not needed for dependency restoration. Compose derives the project name `go-image-optimizer` from this directory; do not add a new project name when working with existing resources.

Inspected API startup reads PORT and starts an HTTP server; no database, migration, seeder or worker exists in this configuration. Frontend API routes call the configured backend on requests. Dependency checks do not run those endpoints. Backend development now runs from mounted source with the same host Go caches used by setup; the production backend remains a compiled Alpine runtime. The production frontend remains image-based.

## Initial setup

```sh
if [ ! -e .env ]; then cp .env.example .env; fi
mkdir -p backend/.cache/go-mod backend/.cache/go-build
docker ps --format '{{.Names}} {{.Ports}}'
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev build backend frontend-dev
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev run --rm --no-deps -T --entrypoint sh frontend-dev -c 'npm ci --no-fund --no-audit'
docker compose -f docker-compose.yml -f compose.development.yaml run --rm --no-deps -T --entrypoint sh backend -c 'go mod download'
docker compose -f docker-compose.yml -f compose.development.yaml run --rm --no-deps -T --entrypoint sh backend -c 'go build -o /tmp/go-image-optimizer-check ./cmd/api'
```

`go mod download` restores modules specified by go.mod/go.sum. `go build` separately resolves compilation needs and creates a Linux executable; it does not launch the server. npm uses `npm ci` and the tracked lockfile. Neither package manager upgrades are needed. One-off containers skip dependencies, publish no ports and bypass normal commands, so installation works even when application containers cannot stay running.

## Physical host storage

| Component / manager | Container path | Host path relative to project | Mount | Purpose |
| --- | --- | --- | --- | --- |
| Next.js / npm | `/app/node_modules` | `frontend/node_modules` | Source bind | Installed packages including Linux native binaries |
| Go modules / Go tools | `/go/pkg/mod` | `backend/.cache/go-mod` | Bind | Downloaded module sources and download metadata |
| Go compiler / Go tools | `/root/.cache/go-build` | `backend/.cache/go-build` | Bind | Reusable compiled package/build cache |
| Next.js artifacts | `/app/.next` | `frontend/.next` | Source bind | Generated dev/build output, not dependencies |

The module cache is Go's native dependency storage, not a vendor directory. No vendoring was enabled. GOMODCACHE and GOCACHE point at these bind mounts for backend and backend-test with the override. The Node download cache `/root/.npm` is disposable container storage; `node_modules` is the required installed tree. System codec libraries stay in the image and are not project package caches. Next.js `.next`, Go executables, coverage and TypeScript incremental files are generated artifacts. Caches/artifacts are ignored by Git; backend .cache is excluded from the Docker build context.

## Start and stop one project

After installing dependencies and checking host listeners:

```sh
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev up -d backend frontend-dev
# Stop only the services you started:
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev stop frontend-dev backend
```

Do not run bare `up` with the dev profile: that also selects the production frontend. No database data needs initialization. Do not use volume removal or pruning. These startup commands are repository-derived unless explicitly marked tested below.

## Deleted dependency/cache recovery

If your applications are running, stop only your backend/frontend-dev before restoration to avoid concurrent reads/writes. Recreate missing cache directories with `mkdir -p backend/.cache/go-mod backend/.cache/go-build`, then rerun npm ci, go mod download and the focused go build above. Deleting GOCACHE only causes compilation work to be repeated; deleting GOMODCACHE requires downloads. Deleting .next needs no dependency reinstall: Next.js regenerates it on development startup or build.

No image rebuild is needed for missing node_modules or Go caches. Rebuild when the toolchain or native system libraries change. If an existing container has an obsolete mount after directory deletion, recreate only the affected service:

```sh
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev up -d --no-deps --force-recreate backend frontend-dev
```

If Go test dependencies are required, build the setup-compatible test service without executing tests:

```sh
docker compose -f docker-compose.yml -f compose.development.yaml --profile test build backend-test
docker compose -f docker-compose.yml -f compose.development.yaml --profile test run --rm --no-deps -T --entrypoint sh backend-test -c 'go mod download'
```

Test execution itself is outside this dependency check. Production runtime containers do not contain the Go compiler/full npm development installation; use the specified development services.

## Missing configuration and generated directories

For missing `.env`, use the guarded template copy above. This restores example defaults, not deleted custom URLs/ports or credentials; custom settings require a backup. Configuration inside `backend/internal/config` and frontend next/TypeScript/postcss files is tracked source, not generated by installation. For any missing tracked file, confirm it is absent and restore that specific path with `git restore --source=HEAD -- path/to/missing-file`; this also applies to Dockerfiles, Compose files, go.mod/go.sum, manifests/lockfiles and tracked `frontend/next-env.d.ts`. Do not restore whole directories over existing edits.

Next.js `.next` contains generated types/build metadata and is regenerated by `npm run dev` or `npm run build` inside `/app`. Those are application startup/build commands, not dependency installers. `next-env.d.ts` is tracked here: recover it from Git before resorting to Next.js regeneration. Custom files mistakenly saved inside generated output or caches require backups; neither npm nor Go recreates custom configuration. No other generated application configuration directory was identified. `storage/testdata` is tracked test input, not a dependency cache; restore missing fixtures individually from Git.

## Troubleshooting and platforms

Use both Compose files consistently, and inspect mounts with `docker inspect CONTAINER --format '{{json .Mounts}}'`. The former named node_modules and .next mounts were removed from frontend-dev, so the source bind is now visible on the host. Existing old containers retain old mounts until recreated; existing named volumes were not deleted. Never copy dependencies to the host while continuing to use such a hidden volume.

For occupied ports, inspect Docker containers and host listeners; choose unused BACKEND_PORT/FRONTEND_PORT values in local .env rather than stopping unrelated services. BACKEND_URL normally remains the container URL, not the host port. On Linux, development images use root: choose a matching host UID/GID for setup when needed, with writable HOME/cache locations and `--user "$(id -u):$(id -g)"`. Keep GOMODCACHE/GOCACHE at their mounted paths; npm can use `-e HOME=/tmp`. Correct only the affected generated directory's ownership if needed, with explicit ownership information; do not apply broad recursive changes or chmod 777.

CGO builds require the image's build-base, pkgconf, libheif-dev, decoder and encoder libraries. Linux/ARM64 or Linux/AMD64 native packages are not macOS/Windows executables; use them in matching containers. Changing container architecture/toolchain may require reinstalling npm native packages and regenerating Go build cache. Platform-specific generated artifacts are not portable backups.

## Verification (2026-10-06)

Originally frontend-dev used a named volume masking host node_modules, and the backend downloaded modules during image builds without host cache binds. Inspected the resolved configurations for all four services. Built development toolchain images, leaving final production targets intact. Installation and compilation results are recorded below. No pre-existing containers were running; no conflicting Docker/host listeners on 8080 or 3000 were observed. No tests, database operations or external requests were executed as application checks.

`npm ci` installed 47 packages into host frontend/node_modules; Next.js resolved at `/app/node_modules/next/package.json` in both the installation container and a fresh container. Observed versions: Node v24.21.0, npm 11.19.0, Go 1.27.1 linux/arm64. Go module download and API build succeeded; a fresh container built the API with GOPROXY=off/GOSUMDB=off using the same persisted caches. A broader offline `go list -m all` failed because some module-graph metadata was not cached; complete offline coverage for every transitive/test module is unverified. Focused offline API compilation passed. Recovery test results are recorded below. Full application startup and test execution were not performed.

Empty-directory recovery passed for npm and Go: separate temporary host binds were mounted over node_modules, GOMODCACHE and GOCACHE; npm ci, go mod download and API compilation succeeded. Physical host Next.js files, Go download ZIPs and build-cache files were confirmed. The temporary directory was removed after the one-off containers exited. Normal dependency/cache directories were never deleted or renamed. No audit containers remain running.

## Copyable host verification and tracked configuration recovery

```sh
# From the host project root.
test -f frontend/node_modules/next/package.json
test -f backend/.cache/go-mod/cache/download/golang.org/x/image/@v/v0.46.0.mod
for path in .env.example docker-compose.yml backend/Dockerfile backend/go.mod backend/go.sum frontend/Dockerfile frontend/package.json frontend/package-lock.json frontend/next.config.ts frontend/next-env.d.ts frontend/tsconfig.json frontend/postcss.config.mjs; do
  if [ ! -e "$path" ]; then git restore --source=HEAD -- "$path"; fi
done
# Generate missing .next output after dependency restoration (build command, not installer):
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev run --rm --no-deps -T --entrypoint npm frontend-dev run build
```

The .next generation command is repository-derived and was not executed in this audit. Back up custom settings before repairing configuration; the guarded Git loop only handles absent tracked files. The new development override and guide require a working-tree backup until committed.
