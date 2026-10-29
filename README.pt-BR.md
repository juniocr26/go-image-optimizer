# go-image-optimizer

[English](README.md) | [Português brasileiro](README.pt-BR.md)

Aplicação Go e Next.js para compressão, redimensionamento, conversão de formatos e fallback de prévia. O processamento ocorre por requisição, sem banco nem persistência de uploads.

## Catálogo de documentação e percurso de leitura

- [Catálogo completo por categoria](docs/pt-br/index.md)
- [Checklist da revisão e lacunas](docs/pt-br/operations/completion-checklist.md)

Leia propósito/setup, arquitetura e contratos, segurança/falhas, testes/evidência e implantação/recuperação local. Este projeto é independente no portfólio; procedimentos runtime documentados não foram executados na revisão de 2026-10-10.

## Documentação

- [001 native image codecs](docs/pt-br/adr/001-native-image-codecs.md)
- [Arquitetura](docs/pt-br/architecture/overview.md)
- [Dependências de desenvolvimento e recuperação](docs/pt-br/docker/development.md)
- [Docker e configuração](docs/pt-br/docker/runtime.md)
- [Guia do projeto](docs/pt-br/guides/project-guide.md)
- [Testes](docs/pt-br/testing/strategy.md)
- [Resultados de validação](docs/pt-br/testing/verification.md)

[Licença](LICENSE)

## Início rápido e escopo

Instale Docker com Compose v2 e execute neste repositório:

```sh
docker compose up --build backend frontend
```

UI: http://localhost:3000; API: http://localhost:8080. Dockerfiles constroem codecs nativos. Não há banco ou histórico persistente de uploads.

[Instalação, fluxo de execução, limites operacionais e entrevista do projeto](docs/pt-br/guides/implementation-study.md). Comandos são orientação; esta manutenção documental não executou testes da aplicação.
