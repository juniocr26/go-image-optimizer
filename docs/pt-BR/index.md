# Catálogo de documentação: go-image-optimizer

[Introdução do projeto](../../README.pt-BR.md) | [Outro idioma](../en/index.md)

Percurso: propósito/setup → arquitetura/contratos → segurança/falhas → testes/evidência → operações. Na biblioteca: catálogo/dossiê → entrevista por projeto → fundamentos → falhas. Marcos/refactors/handoff históricos descrevem datas originais; revisão/checklist explica escopo atual.

## architecture

Componentes, fluxos e fronteiras implementadas.

- [Arquitetura](../pt-br/architecture/overview.md)

## adr

Decisões e consequências; IDs/histórico preservados.

- [ADR 001: Codecs nativos de imagem](../pt-br/adr/001-native-image-codecs.md)

## guides

Setup, pré-requisitos e procedimentos de leitura.

- [Go Image Optimizer](../pt-br/guides/project-guide.md)

## testing

Estratégia, inventários e evidências datadas.

- [Documentação de Testes](../pt-br/testing/strategy.md)
- [Validação documental — 2026-10-06](../../../engineering-library/docs/engineering-library/pt-br/testing/documentation.md)

## docker

Imagens, serviços, mounts e configuração local.

- [Desenvolvimento Docker e recuperação](../pt-br/docker/development.md)
- [Docker](../pt-br/docker/runtime.md)

## api

Contratos expostos e comportamento de clientes.

- [API de imagens e contrato do proxy](../pt-br/api/contracts.md)

## operations

Diagnóstico, recuperação, checklists e registros históricos.

- [go-image-optimizer: Checklist de conclusão do repositório](../pt-br/operations/completion-checklist.md)
- [Implantação, diagnóstico e evidência](../pt-br/operations/deployment-and-diagnosis.md)

## security

Autenticação, autorização, dados e recursos.

- [Limites de recursos e código nativo](../pt-br/security/resource-boundaries.md)

Aplicabilidade e omissões estão justificadas no [checklist](../pt-br/operations/completion-checklist.md). Não há categorias vazias.

## Roteiro prático de implementação e entrevista

[Acompanhe uma requisição, examine falhas e pratique respostas](../pt-br/guides/implementation-study.md).

- [Validação documental — 2026-10-06](testing/verification.md)
