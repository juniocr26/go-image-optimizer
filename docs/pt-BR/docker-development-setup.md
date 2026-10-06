# Desenvolvimento Docker e recuperação

Execute no diretório `go-image-optimizer/` do host. São necessários Docker Engine/Desktop, Compose com múltiplos arquivos e profiles (auditoria anterior: v5.1.4), acesso aos registries npm/Go/Alpine e memória suficiente para build. Go/npm no host não são necessários. Backend: Go 1.27.1 Alpine com CGO; frontend: tag flutuante `node:24-alpine`. Versões frontend instaladas vêm do lockfile, não das faixas em package.json.

## Serviços e inicialização

`backend` oferece API na 8080; `frontend` é Next.js standalone de produção na 3000; `frontend-dev` (profile `dev`) usa a mesma porta, portanto não execute os dois frontends juntos. `backend-test` (profile `test`) executa testes Go e monta fixtures somente para leitura; não é necessário para recuperar dependências. Compose deriva `go-image-optimizer` do diretório; preserve o nome ao usar recursos existentes.

O startup da API lê PORT e inicia servidor HTTP. Não há banco, migração, seed ou worker. Rotas API frontend chamam backend por requisição. Verificações de dependências não executam endpoints. Backend development usa fontes montadas e caches Go do host; produção continua binário Alpine. Frontend de produção continua baseado na imagem.

## Instalação inicial

```sh
if [ ! -e .env ]; then cp .env.example .env; fi
mkdir -p backend/.cache/go-mod backend/.cache/go-build
docker ps --format '{{.Names}} {{.Ports}}'
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev build backend frontend-dev
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev run --rm --no-deps -T --entrypoint sh frontend-dev -c 'npm ci --no-fund --no-audit'
docker compose -f docker-compose.yml -f compose.development.yaml run --rm --no-deps -T --entrypoint sh backend -c 'go mod download'
docker compose -f docker-compose.yml -f compose.development.yaml run --rm --no-deps -T --entrypoint sh backend -c 'go build -o /tmp/go-image-optimizer-check ./cmd/api'
```

`go mod download` recupera módulos de go.mod/go.sum. `go build` resolve compilação e cria executável Linux sem iniciar servidor. npm ci usa lockfile; não são necessários upgrades. Containers one-off ignoram dependências, não publicam portas e substituem comandos normais; instalação funciona mesmo se a aplicação não permanecer ativa.

## Armazenamento físico no host

| Componente | Caminho no container | Caminho no host | Montagem / finalidade |
| --- | --- | --- | --- |
| Next.js/npm | `/app/node_modules` | `frontend/node_modules` | Bind das fontes; pacotes e binários Linux |
| Módulos Go | `/go/pkg/mod` | `backend/.cache/go-mod` | Bind; fontes e metadados de downloads |
| Compilador Go | `/root/.cache/go-build` | `backend/.cache/go-build` | Bind; cache de pacotes compilados |
| Artefatos Next.js | `/app/.next` | `frontend/.next` | Bind; saída dev/build, não dependências |

Go usa cache nativo, sem vendor. GOMODCACHE/GOCACHE apontam aos binds de backend/backend-test no override. `/root/.npm` é cache descartável; node_modules é a instalação. Codecs de sistema ficam na imagem, não no cache de pacotes. `.next`, executáveis Go, cobertura e arquivos incrementais TypeScript são gerados. Git ignora caches/artefatos; backend .cache fica fora do contexto Docker.

## Iniciar e parar um projeto

Após instalar e conferir listeners:

```sh
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev up -d backend frontend-dev
# Pare somente serviços que iniciou:
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev stop frontend-dev backend
```

Não use `up` sem serviços com profile dev: isso seleciona frontend de produção também. Não há inicialização de banco. Não remova volumes nem execute pruning. Os comandos foram derivados do projeto; resultados atuais ficam em [verificação](verification.md).

## Recuperação de dependências e caches excluídos

Pare somente backend/frontend-dev antes de recuperar para evitar acesso simultâneo. Recrie diretórios com `mkdir -p backend/.cache/go-mod backend/.cache/go-build` e repita npm ci, go mod download e go build. Excluir GOCACHE exige recompilar; excluir GOMODCACHE exige downloads. Next.js regenera .next ao iniciar/build sem reinstalar dependências.

Não é preciso rebuild por caches/node_modules ausentes; rebuild quando toolchain/bibliotecas nativas mudarem. Para mount obsoleto após exclusão de diretório, recrie somente serviço afetado:

```sh
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev up -d --no-deps --force-recreate backend frontend-dev
```

Para dependências de teste sem executar testes:

```sh
docker compose -f docker-compose.yml -f compose.development.yaml --profile test build backend-test
docker compose -f docker-compose.yml -f compose.development.yaml --profile test run --rm --no-deps -T --entrypoint sh backend-test -c 'go mod download'
```

Execução de testes fica fora da verificação de dependências. Runtime de produção não tem compilador Go/instalação npm completa; use serviços development.

## Configuração ausente e diretórios gerados

Copie `.env.example` apenas se `.env` não existir; defaults não recuperam URLs, portas ou credenciais personalizadas, que exigem backup. `backend/internal/config` e configuração next/TypeScript/postcss são fontes versionadas. Confirme ausência e restaure caminho específico com `git restore --source=HEAD -- path/to/missing-file`; aplica-se a Dockerfiles, Compose, go.mod/go.sum, manifests/lockfiles e `frontend/next-env.d.ts`. Não restaure diretórios inteiros sobre alterações existentes.

`.next` contém tipos/metadados gerados por `npm run dev`/`npm run build` em `/app`, comandos de execução/build, não instaladores. `next-env.d.ts` é versionado: restaure do Git antes de regenerar. Arquivos personalizados guardados em caches/saídas exigem backup; npm/Go não os recriam. Nenhum outro diretório gerado de configuração foi identificado. `storage/testdata` é entrada de teste versionada, não cache; restaure fixtures ausentes individualmente.

## Diagnóstico e plataformas

Use ambos os arquivos e confira mounts com `docker inspect CONTAINER --format '{{json .Mounts}}'`. Mounts nomeados antigos de node_modules/.next foram removidos; bind agora aparece no host. Containers antigos mantêm mounts até recriação; volumes existentes não foram excluídos. Não copie dependências ao host enquanto um volume oculto continuar em uso.

Em conflito de porta, inspecione Docker/listeners do host e configure BACKEND_PORT/FRONTEND_PORT livres no .env local sem parar serviços alheios. BACKEND_URL normalmente mantém URL do container, não porta do host. Imagens development usam root; no Linux, use UID/GID do host quando necessário, HOME/cache graváveis e `--user "$(id -u):$(id -g)"`. Preserve caminhos montados GOMODCACHE/GOCACHE; npm pode usar `-e HOME=/tmp`. Corrija somente propriedade do diretório gerado afetado com informação explícita; sem mudanças recursivas amplas ou chmod 777.

CGO exige build-base, pkgconf, libheif-dev e plugins decoder/encoder. Pacotes Linux ARM64/AMD64 não são executáveis macOS/Windows; use containers compatíveis. Alterar arquitetura/toolchain pode exigir reinstalação npm nativa e regeneração GOCACHE. Artefatos de plataforma não são backups portáveis.

## Auditoria anterior — 2026-10-06

Frontend-dev antes usava volume nomeado ocultando node_modules; backend baixava módulos no build sem bind de cache. Configuração dos quatro serviços inspecionada; imagens development construídas preservando targets de produção. Não havia containers rodando nem conflitos observados na 8080/3000. Nenhum teste, banco ou requisição externa como validação da aplicação foi executado naquela auditoria.

npm ci instalou 47 pacotes no host; Next.js resolveu `/app/node_modules/next/package.json` na instalação e em container novo. Versões observadas: Node v24.21.0, npm 11.19.0, Go 1.27.1 linux/arm64. Download/build API passaram; container novo compilou offline com GOPROXY=off/GOSUMDB=off. `go list -m all` offline falhou por metadados transitivos ausentes; cobertura offline completa não verificada. Compilação focada passou. Startup completo e testes não foram realizados naquela auditoria.

Recuperação de diretório vazio passou para npm/Go com binds temporários separados sobre node_modules/GOMODCACHE/GOCACHE. npm ci, download/build passaram; arquivos Next.js, ZIPs Go e cache físicos conferidos. Temporário removido após containers encerrados; diretórios normais nunca excluídos/renomeados. Nenhum container de auditoria ficou ativo.

## Conferência física e recuperação versionada

```sh
test -f frontend/node_modules/next/package.json
test -f backend/.cache/go-mod/cache/download/golang.org/x/image/@v/v0.46.0.mod
for path in .env.example docker-compose.yml backend/Dockerfile backend/go.mod backend/go.sum frontend/Dockerfile frontend/package.json frontend/package-lock.json frontend/next.config.ts frontend/next-env.d.ts frontend/tsconfig.json frontend/postcss.config.mjs; do
  if [ ! -e "$path" ]; then git restore --source=HEAD -- "$path"; fi
done
# Gerar .next após recuperar dependências; build, não instalador:
docker compose -f docker-compose.yml -f compose.development.yaml --profile dev run --rm --no-deps -T --entrypoint npm frontend-dev run build
```

Esse comando de geração foi derivado do projeto e não executado na auditoria anterior. Faça backup de configurações personalizadas antes de reparar; o loop Git só trata arquivos ausentes. Confira versionamento do override/guia e preserve alterações locais em backup.
