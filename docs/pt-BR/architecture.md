# Arquitetura

Este documento descreve a arquitetura atual, os trade-offs e a evolução esperada do Go Image Optimizer.

O projeto evolui de forma incremental. Novos componentes e padrões só devem ser introduzidos quando um requisito concreto ou uma limitação observada justificar essa complexidade.

## 1. Contexto

Go Image Optimizer é uma aplicação para otimização de imagens com backend em Go e interface web construída com Next.js, React e Tailwind CSS.

A implementação atual entrega um fluxo síncrono de compressão exatamente para estes formatos:

- JPEG / JPG
- PNG
- WebP
- AVIF
- HEIC / HEIF
- GIF
- BMP
- TIFF

WebM, SVG, formatos RAW de câmera, vídeos, arquivos compactados e formatos arbitrários de imagem não são suportados.

## 2. Fluxo atual da requisição

```mermaid
sequenceDiagram
    actor User as Usuário
    participant Browser as Interface no navegador
    participant NextAPI as Rota API do Next.js
    participant Handler as Handler HTTP em Go
    participant UseCase as Caso de uso de compressão
    participant Compressor as Implementação de compressão

    User->>Browser: Seleciona uma imagem suportada
    Browser->>Browser: Cria uma URL Blob temporária quando o navegador consegue renderizar
    User->>Browser: Clica em Compress
    Browser->>NextAPI: POST /api/images/compress
    NextAPI->>Handler: POST /images/compress
    Handler->>Handler: Valida multipart e limite de upload
    Handler->>UseCase: Executa a compressão com os bytes da imagem
    UseCase->>Compressor: Comprime com base no formato detectado pelos bytes
    Compressor-->>UseCase: Retorna bytes otimizados e metadados
    UseCase-->>Handler: Retorna o resultado
    Handler-->>NextAPI: Retorna os bytes da imagem otimizada
    NextAPI-->>Browser: Repassa bytes e headers da resposta
    Browser->>Browser: Cria uma URL Blob temporária para o resultado
    Browser-->>User: Exibe medições reais e ação de download
```

O navegador mantém o preview da imagem selecionada e o resultado comprimido apenas em estado React e URLs Blob. Essas URLs são revogadas quando são substituídas, quando o fluxo é reiniciado ou quando o componente é desmontado. Ao recarregar a página, a sessão atual desaparece de forma intencional.

## 3. Fronteiras no backend

O backend possui uma fronteira pequena de aplicação:

```text
Handler HTTP
    -> Caso de uso de compressão de imagem
        -> Implementação de compressão de imagem
```

Responsabilidades atuais:

- Handler HTTP: leitura do multipart, limite de 25 MiB, validação dos campos, status codes, headers de resposta e geração do nome de download.
- Caso de uso de compressão: execução da regra de aplicação e checagens de contexto, sem depender de HTTP ou tipos de multipart.
- Implementação de compressão: identificação do conteúdo real pelos bytes, validação da imagem, proteção por dimensão e por animação, decode, encode específico por formato e configurações de compressão.

O projeto ainda não cria um modelo de domínio porque a funcionalidade atual não possui entidades de domínio relevantes. A interface do compressor existe como uma fronteira útil entre o caso de uso e a implementação de infraestrutura.

## 4. Detecção de formato

O backend não confia na extensão do arquivo nem no MIME type informado pelo navegador. Ele inspeciona os bytes enviados e aceita apenas assinaturas e marcas de container conhecidas para os formatos suportados.

Por isso, um JPEG enviado como `sample.png` ainda é processado como JPEG e retorna com extensão de download JPEG. Um arquivo chamado `broken.webp` com bytes que não são WebP é rejeitado em vez de ser encaminhado ao codec WebP.

## 5. Comportamento da compressão

A aplicação preserva a família do formato de origem para imagens suportadas. Ela não faz conversão visível entre formatos não relacionados.

Comportamento atual dos codecs:

- JPEG é decodificado, a orientação EXIF é aplicada aos pixels e a imagem é reencodada como JPEG com qualidade `82`.
- PNG é decodificado e reencodado como PNG com o melhor nível de compressão da biblioteca padrão. Pixels e alpha são lossless.
- WebP estático é decodificado e reencodado como WebP. Entrada WebP lossless continua lossless; alpha é preservado.
- WebP animado é decodificado pelo container de animação, reconstruído como frames de canvas completo e reencodado como WebP animado. Quantidade de frames, duração, loop, cor de fundo e chunks de metadados suportados são preservados, mas retângulos internos de sub-frame e escolhas de disposal podem ser normalizados pelo encoder.
- AVIF é decodificado e reencodado como AVIF. A rotação automática no decode fica habilitada. AVIF com múltiplos frames é codificado com delays e loop quando o codec consegue decodificar.
- HEIC/HEIF usa libheif e suporte HEVC nativos. O backend aceita uma imagem primária top-level e rejeita variantes multi-imagem não suportadas com `422`.
- GIF é decodificado com a biblioteca padrão do Go e reencodado como GIF. Frames, delays, disposal e loop de GIF animado são preservados por `gif.EncodeAll`.
- BMP é decodificado e reencodado como BMP. A saída BMP pode não ficar menor.
- TIFF é decodificado e reencodado como TIFF com compressão Deflate e predictor habilitado.

Arquivos já otimizados podem continuar com o mesmo tamanho ou ficar maiores. O frontend exibe medições reais de bytes em vez de presumir redução.

## 6. Ciclo de vida dos arquivos e armazenamento

O ciclo de vida atual no backend é efêmero:

```text
Navegador
    -> POST da imagem
    -> Go recebe os bytes
    -> Go comprime os bytes
    -> Go retorna os bytes otimizados
    -> Navegador mantém o resultado temporariamente
    -> Usuário baixa o resultado
```

Imagens enviadas e imagens comprimidas não são persistidas em armazenamento da aplicação. O backend não cria IDs de processamento, registros em banco de dados, registros no Redis, objetos em storage, URLs de resultado, filas, jobs em background, histórico de processamento ou limpeza por TTL.

Arquivos temporários de multipart, caso a biblioteca padrão crie algum durante o parsing da requisição, são removidos com `MultipartForm.RemoveAll()` antes do fim da requisição.

A codificação HEIC/HEIF usa internamente a API de saída para arquivo do binding Go da libheif. O compressor escreve em um arquivo temporário do sistema operacional, lê o resultado de volta para memória e remove esse arquivo temporário antes de devolver a resposta. Isso não cria armazenamento durável da aplicação.

## 7. Processamento síncrono

Hoje a compressão roda de forma síncrona dentro da requisição HTTP em Go porque a aplicação devolve um download imediato.

A aplicação não declara características de alta vazão ou escalabilidade. Qualquer afirmação desse tipo precisa ser medida em cargas realistas antes de entrar na documentação.

Proteções atuais de recursos:

- O corpo da requisição é limitado a 25 MiB.
- O parsing multipart mantém até 8 MiB em memória antes de a biblioteca padrão poder usar arquivos temporários.
- As dimensões decodificadas são limitadas a 32 megapixels.
- GIF animado, WebP animado e AVIF com múltiplos frames são limitados por `largura * altura * quantidade de frames`, com limite padrão de 64 milhões de pixels de canvas-frame.

## 8. Ciclo de vida no frontend

O frontend mantém o fluxo em estado React:

- drop zone inicial;
- preview da imagem selecionada quando o navegador consegue renderizar o formato;
- placeholder sem preview para formatos que muitos navegadores não renderizam, como HEIC ou TIFF;
- tamanho original do arquivo;
- ação explícita de Compress;
- estado de carregamento indeterminado;
- preview do resultado quando o navegador consegue renderizar;
- medições reais em bytes, cálculo de redução e ação de download;
- reinício do fluxo para outra imagem.

A interface não persiste a sessão em `localStorage`, IndexedDB, armazenamento do backend ou qualquer outro armazenamento durável. Após recarregar a página, a imagem selecionada e o resultado desaparecem por decisão do MVP.

## 9. Decisão sobre codecs nativos

O suporte a HEIC/HEIF exige libheif nativa e plugins de codec HEVC no Docker. Por isso, a imagem do backend usa build Alpine com CGO habilitado e runtime Alpine com pacotes libheif, em vez de uma imagem distroless totalmente estática.

Consulte [ADR 001: Codecs nativos de imagem](adr-001-codecs-nativos.md) e [Docker](docker.md).

## 10. Limitações atuais

- A compressão é síncrona.
- Preservação de metadados é best-effort e específica por formato, não uma garantia universal.
- Variantes não suportadas são rejeitadas em vez de aproximadas.
- Algumas saídas podem ter o mesmo tamanho ou ficar maiores que o arquivo enviado.
- O suporte de preview no navegador varia por formato.
- Não há histórico, busca por ID, worker em background, fila, banco de dados, object storage ou limpeza por TTL.

## 11. Possível evolução

Se requisitos futuros exigirem processamento assíncrono, arquivos maiores, formatos mais pesados, maior vazão, compartilhamento de resultados ou histórico, a arquitetura pode evoluir para algo como:

```text
Upload
    -> ID de processamento
    -> Fila / worker
    -> Armazenamento temporário ou object storage
    -> Recuperação do resultado
    -> Limpeza por TTL
```

Essa direção deve ser introduzida somente com requisitos claros e trade-offs documentados sobre armazenamento, retenção, limpeza, observabilidade, segurança e custo operacional.
