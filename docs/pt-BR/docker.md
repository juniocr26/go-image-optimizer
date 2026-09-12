# Docker

A configuração Docker executa o mesmo fluxo síncrono e sem armazenamento persistente da aplicação local.

## Serviços

- `backend`: API Go na porta `8080`.
- `frontend`: interface Next.js na porta `3000`.

O `docker-compose.yml` não monta volume de armazenamento da aplicação para imagens enviadas ou otimizadas. As imagens são recebidas, processadas, devolvidas e descartadas.

## Imagem do backend

A imagem do backend usa build Alpine com CGO habilitado porque o suporte a HEIC/HEIF depende da libheif nativa e de plugins de codec HEVC. O estágio de runtime fica fixado em Alpine 3.24 para acompanhar a divisão de pacotes de plugin libheif usada pela imagem Alpine do Go.

Pacotes no estágio de build:

- `build-base`
- `pkgconf`
- `libheif-dev`
- `libheif-libde265`
- `libheif-x265`

Pacotes no runtime:

- `libheif`
- `libheif-libde265`
- `libheif-x265`

O runtime distroless totalmente estático usado antes não é adequado para este conjunto de formatos porque a libheif carrega bibliotecas/plugins nativos em tempo de execução.

## Comandos de validação

```bash
docker compose config
docker compose build
docker compose up
```

Depois, abra o frontend em `http://localhost:3000` e envie amostras representativas de JPEG/JPG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP e TIFF.

Para executar apenas os testes do backend via Docker:

```bash
docker run --rm -v "$PWD/backend:/src" -w /src golang:1.27.1-alpine sh -lc \
  'apk add --no-cache build-base pkgconf libheif-dev libheif-libde265 libheif-x265 >/dev/null && /usr/local/go/bin/go test ./...'
```

## Diagnóstico

Se a codificação HEIC/HEIF falhar com erro de codec ou plugin, confirme que a imagem de runtime inclui `libheif-x265`. Algumas distribuições Linux separam a libheif em pacotes distintos para decoder e encoder.

Se a divisão de pacotes da distribuição for diferente da Alpine, instale os pacotes equivalentes de runtime libheif, decoder HEIC e encoder HEVC. `libheif-plugins-all` pode ajudar em diagnóstico, mas o Dockerfile do projeto mantém os pacotes de runtime mais restritos.
