# Go Image Optimizer

Aplicação para otimização de imagens desenvolvida em Go, com uma interface web utilizando Next.js, React e Tailwind CSS.

O projeto será desenvolvido de forma incremental, começando pela compressão de imagens e evoluindo sua arquitetura conforme novos requisitos e desafios técnicos surgirem.

> **Status atual:** Início do desenvolvimento. A primeira funcionalidade planejada é a compressão de imagens.

## Visão geral

O Go Image Optimizer é um projeto de portfólio voltado ao processamento de imagens e à aplicação de conceitos de engenharia de backend utilizando Go.

Em vez de definir uma arquitetura complexa antecipadamente, o projeto segue uma abordagem incremental: começar com uma solução simples, validar os requisitos e introduzir mudanças arquiteturais quando existir uma razão concreta para isso.

## Escopo inicial

A primeira funcionalidade permitirá:

- Enviar uma imagem pela interface web.
- Enviar essa imagem para o backend em Go.
- Comprimir a imagem.
- Receber a imagem comprimida como resultado.

Outras funcionalidades de otimização serão adicionadas gradualmente conforme o projeto evoluir.

## Tecnologias

### Backend

- Go

### Frontend

- Next.js
- React
- Tailwind CSS

## Arquitetura

A arquitetura inicial mantém, propositalmente, o processamento da imagem dentro da aplicação Go.

```mermaid
flowchart LR
    U[Usuário] --> F[Interface Web - Next.js]
    F -->|Upload da imagem| API[Aplicação Go]
    API --> C[Compressão da imagem]
    C --> API
    API -->|Imagem comprimida| F
    F --> U
```

Essa arquitetura é intencionalmente simples. Novos componentes ou serviços serão introduzidos somente quando requisitos ou limitações observadas justificarem a complexidade adicional.

Para conhecer as decisões arquiteturais e seus trade-offs, consulte [Arquitetura](docs/architecture.pt-BR.md).

## Documentação

- [Architecture — English](docs/architecture.md)
- [Arquitetura](docs/architecture.pt-BR.md)
- [README — English](README.md)

## Roadmap

O projeto será desenvolvido de forma incremental. Entre as funcionalidades planejadas estão:

- Compressão de imagens
- Redimensionamento
- Conversão de formatos
- Geração de WebP e AVIF
- Geração de thumbnails
- Comparação entre o tamanho original e otimizado
- Histórico de processamentos

O roadmap representa a direção pretendida para o projeto e poderá mudar conforme as decisões de implementação e os requisitos técnicos evoluírem.

## Licença

Nenhuma licença foi definida para este projeto até o momento.
