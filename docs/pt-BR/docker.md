# Docker

A configuração Docker executa o mesmo fluxo síncrono e sem armazenamento persistente da aplicação local.

## Serviços

- `backend`: API Go na porta `8080`.
- `frontend`: interface Next.js na porta `3000`.
- `backend-test`: executor de testes do backend, protegido por profile, usando o estágio de build Go.

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

O serviço de produção `backend` usa o estágio final de runtime e não inclui o toolchain Go nem código-fonte montado. Os testes do backend rodam pelo serviço separado `backend-test`, que usa o estágio de build, mantém CGO e dependências nativas de codec disponíveis, e monta `./backend` em `/src`.

## Comandos de validação

```bash
docker compose config
docker compose build backend
docker compose build backend-test
docker compose up
```

Depois, abra o frontend em `http://localhost:3000` e envie amostras representativas de JPEG/JPG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP e TIFF.

Para executar apenas os testes do backend via Docker:

```bash
docker compose run --build --rm backend-test
```

O serviço de teste fica atrás do profile `test`, então ele não é iniciado pelo `docker compose up` normal. Para passar flags customizadas do Go test no mesmo ambiente com codecs nativos:

```bash
docker compose run --build --rm backend-test go test -v ./internal/infrastructure/imaging -run TestCompressorCompressesSupportedStaticFormats
```

`go test -race ./...` está bloqueado no momento por uma falha de `checkptr` do Go dentro de `github.com/strukturag/libheif` durante a geração das fixtures HEIC/HEIF. A suíte normal sem `-race` é o fluxo de testes do backend suportado.

## Diagnóstico

Se a codificação HEIC/HEIF falhar com erro de codec ou plugin, confirme que a imagem de runtime inclui `libheif-x265`. Algumas distribuições Linux separam a libheif em pacotes distintos para decoder e encoder.

Se a divisão de pacotes da distribuição for diferente da Alpine, instale os pacotes equivalentes de runtime libheif, decoder HEIC e encoder HEVC. `libheif-plugins-all` pode ajudar em diagnóstico, mas o Dockerfile do projeto mantém os pacotes de runtime mais restritos.
