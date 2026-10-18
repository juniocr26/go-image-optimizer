# go-image-optimizer

[English](README.md) | [Português brasileiro](README.pt-BR.md)

Aplicação Go e Next.js para compressão, redimensionamento, conversão de formatos e fallback de prévia. O processamento ocorre por requisição, sem banco nem persistência de uploads.

## Catálogo de documentação e percurso de leitura

- [Catálogo completo por categoria](docs/pt-BR/index.md)
- [Checklist da revisão e lacunas](docs/pt-BR/operations/completion-checklist.md)

Leia propósito/setup, arquitetura e contratos, segurança/falhas, testes/evidência e implantação/recuperação local. Este projeto é independente no portfólio; procedimentos runtime documentados não foram executados na revisão de 2026-10-10.

## Documentação

- [001 native image codecs](docs/pt-BR/adr/001-native-image-codecs.md)
- [Arquitetura](docs/pt-BR/architecture/overview.md)
- [Dependências de desenvolvimento e recuperação](docs/pt-BR/docker/development.md)
- [Docker e configuração](docs/pt-BR/docker/runtime.md)
- [Guia do projeto](docs/pt-BR/guides/project-guide.md)
- [Testes](docs/pt-BR/testing/strategy.md)
- [Resultados de validação](docs/pt-BR/testing/verification.md)

[Licença](LICENSE)

## Início rápido e escopo

Instale Docker com Compose v2 e execute neste repositório:

```sh
docker compose up --build backend frontend
```

UI: http://localhost:3000; API: http://localhost:8080. Dockerfiles constroem codecs nativos. Não há banco ou histórico persistente de uploads.

[Instalação, fluxo de execução, limites operacionais e entrevista do projeto](docs/pt-BR/guides/implementation-study.md). Comandos são orientação; esta manutenção documental não executou testes da aplicação.
