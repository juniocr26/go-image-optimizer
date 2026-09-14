# ADR 001: Codecs nativos de imagem

## Status

Aceito.

## Contexto

A aplicação agora suporta HEIC/HEIF além dos formatos que podem ser tratados pela biblioteca padrão do Go ou por codecs pure-Go.

O suporte a HEIC/HEIF não existe de forma prática na biblioteca padrão do Go. A implementação escolhida usa o binding Go da libheif. Isso entrega decode e encode reais de HEIC/HEIF, mas exige CGO e bibliotecas/plugins nativos da libheif no build e no runtime.

## Decisão

Usar libheif nativa para suporte a HEIC/HEIF e manter o restante do pipeline síncrono e limitado ao escopo da requisição.

A imagem Docker do backend usa:

- estágio de build Alpine com `CGO_ENABLED=1`;
- `libheif-dev`, `libheif-libde265` e `libheif-x265` durante o build;
- runtime Alpine com `libheif`, `libheif-libde265` e `libheif-x265`;
- serviço Compose `backend-test` baseado no estágio de build para testes locais que precisam de Go, CGO e headers nativos de codec;
- nenhum armazenamento durável de imagem, fila, banco de dados ou object store.

## Consequências

Pontos positivos:

- O suporte a HEIC/HEIF é real, não apenas uma promessa baseada em nome de arquivo ou MIME type.
- Variantes HEIF não suportadas podem ser rejeitadas explicitamente.
- O restante da arquitetura da aplicação permanece simples e síncrono.

Trade-offs:

- O backend não pode mais usar um runtime distroless totalmente estático.
- Builds nativos locais precisam das bibliotecas de desenvolvimento equivalentes da libheif instaladas.
- Imagens de runtime precisam incluir os plugins de codec necessários da libheif.
- Testes do backend via Docker devem rodar no serviço de teste baseado no estágio de build, para que fixtures reais HEIC/HEIF e geração sintética de HEIC usem a mesma cadeia de dependências nativas.
- O binding Go da libheif escreve a saída codificada por uma API de arquivo temporário, então o compressor cria e remove imediatamente um arquivo temporário do sistema operacional para a saída HEIC/HEIF.

A decisão deve ser revisitada se requisitos futuros exigirem runtime totalmente estático, suporte mais amplo a variantes HEIF, offload para serviço nativo/GPU ou processamento assíncrono de alta vazão.
