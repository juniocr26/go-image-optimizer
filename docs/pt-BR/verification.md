# Validação documental — 2026-10-06

Não havia containers Docker rodando antes deste projeto. Com `-f docker-compose.yml -f compose.development.yaml`, `--profile dev config --quiet` e `--profile dev up -d backend frontend-dev` passaram. Não há banco, migração, seed, worker nem volume persistente de uploads. `--profile test run --rm --no-deps -T backend-test go test -count=1 ./...` passou todos os pacotes com testes, incluindo HEIF nativo e fixtures reais somente leitura. A imagem development de testes ausente foi construída pelo target existente após tentativa de pull no registry.

`--profile dev exec -T frontend-dev node --test tests/resize-options.test.mjs tests/conversion-size.test.mjs tests/preview-cache.test.mjs` passou todos os 12 testes; Node emitiu avisos de tipo de módulo. `exec -T frontend-dev npx tsc --noEmit` passou. `/health` do backend e `/` do frontend retornaram HTTP 200 dentro do container frontend. curl no host não alcançou a porta publicada no contexto restrito; consulte os resultados em container abaixo. Build de produção, interação no navegador, fidelidade visual, race e carga não foram verificados.

Serviços iniciados aqui foram parados com os mesmos arquivos e `--profile dev stop frontend-dev backend`. Volumes existentes, fixtures e código foram preservados. Next.js regenerou `next-env.d.ts`; o conteúdo versionado original foi restaurado após a parada e o novo arquivo TypeScript build-info não versionado foi removido. Pares de idiomas e links/âncoras locais conferidos.

Smoke checks em container enviaram a fixture PNG existente para `/api/images/resize/info` e `/api/images/preview`: ambos HTTP 200. Inspeção informou 512 × 512, png, image/png e um frame; prévia devolveu 26.968 bytes. A saída permaneceu em memória.
