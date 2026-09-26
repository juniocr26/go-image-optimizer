# Redimensionamento de imagens — Ciclo 2

## Fluxo de uso

Selecione uma imagem, escolha **Resize** e clique em **Run**. O modal de configuração abre sem mudar o layout da Home, consulta as dimensões no backend e mostra o preview original (ou fallback do navegador) e as dimensões previstas da saída.

- **Pixels** começa com as dimensões originais, considerando a orientação de exibição. **Keep aspect ratio** e **Don't enlarge** vêm marcados. A última dimensão editada determina o cálculo proporcional da outra. Destravar permite distorcer a proporção; não há recorte nem preenchimento.
- **Percentage** oferece **25% smaller**, **50% smaller** e **75% smaller**, com 50% selecionado inicialmente. A porcentagem reduz largura e altura, não os bytes nem a área total de pixels.
- O arredondamento é para o inteiro mais próximo, com mínimo de um pixel. Reduzir 899 × 1599 em 50% resulta em 450 × 800.
- Com proporção travada, Don't enlarge limita a escala a 1. Com ela destravada, cada dimensão é limitada individualmente ao valor original. O resumo sempre mostra as dimensões efetivas.
- Desmarcar Don't enlarge permite ampliar dentro dos limites de recursos. A ampliação não recupera detalhes.
- **Resize image** executa a operação. **Cancel**, o botão de fechar e Escape fecham a configuração quando não há processamento. Clicar no fundo não fecha. Durante o processamento, opções e fechamento ficam desabilitados; o carregamento é indeterminado.
- O resultado substitui o modal de configuração e mostra dimensões e tamanhos reais, preview/fallback e download. O modal de resultado mantém o fechamento explícito existente.
- O original continua selecionado para outra operação. Reabrir a configuração restaura os valores iniciais. Erros de processamento mantêm as opções; erros na leitura de dimensões oferecem uma ação de tentar novamente.

O modal de configuração usa dialog nativo para conter o foco e tornar o fundo inativo, foco inicial no título, abas navegáveis por teclado, bloqueio de rolagem do body e restauração do foco. Em telas estreitas, vira uma coluna com conteúdo rolável e rodapé fixo dentro do modal.

## API

As duas rotas recebem `multipart/form-data` com exatamente um arquivo no campo `image`. Formato e dimensões vêm dos bytes reais; MIME do navegador, extensão e dimensões informadas pelo cliente não são fontes de verdade.

### POST /images/resize/info

Retorna JSON, por exemplo:

```json
{"width":899,"height":1599,"format":"jpeg","contentType":"image/jpeg","frameCount":1}
```

A inspeção valida e decodifica a origem, sem recodificar nem persistir. As dimensões consideram a orientação EXIF de JPEG e a orientação nativa suportada. Isso funciona também quando o navegador não consegue gerar preview. A leitura pode custar mais em imagens grandes/codecs nativos; o modal mostra carregamento. Fechar o modal aborta a requisição no navegador.

### POST /images/resize

| Campo | Contrato |
| --- | --- |
| `mode` | Obrigatório: `pixels` ou `percentage` |
| `width`, `height` | Obrigatórios em pixels: inteiros positivos, individualmente até 32.000.000; a saída efetiva também deve respeitar o limite total de pixels |
| `axis` | `width` (padrão) ou `height`: última dimensão editada, usada no cálculo proporcional |
| `keepAspectRatio` | `true` (padrão) ou `false`; aplica-se ao modo pixels |
| `withoutEnlargement` | `true` (padrão) ou `false` |
| `reduction` | Obrigatório em percentage: `25`, `50` ou `75` |

Booleanos usam literalmente `true`/`false`; valores explicitamente vazios são rejeitados. Os padrões valem quando o campo é omitido. Opções repetidas, arquivos adicionais e campos `image` misturando texto e arquivo são rejeitados. A interface envia números válidos como inteiros decimais, inclusive quando digitados em notação exponencial. Percentage ignora largura/altura e sempre preserva proporção. O servidor recalcula a saída a partir do arquivo enviado; dimensões de origem não são controladas pelo cliente. A operação sempre parte do original selecionado, nunca de um resultado anterior.

A resposta de sucesso contém os bytes e:

- `Content-Type`, `Content-Length` e `Content-Disposition` com nome sanitizado `*_resized` e extensão da família real;
- `X-Original-Width`, `X-Original-Height`, `X-Image-Width`, `X-Image-Height`, em pixels com orientação de exibição;
- `Cache-Control: no-store`.

Erros retornam JSON `{ "error": "..." }`: 400 para input/opções inválidos, 413 para limites, 415 para formato não suportado, 422 para variante não suportada e 500 para falhas internas/de codec. O proxy Next.js retorna 502 quando não consegue acessar o backend.

As rotas de mesma origem no Next.js são `/api/images/resize/info` e `/api/images/resize`. Um helper de encaminhamento, compartilhado com compressão, repassa multipart, status, MIME, nome e headers de dimensões.

## Arquitetura e ciclo de vida

```text
ImageUploadForm → ImageResizeModal
  → rota Next.js → handler resize_image
    → caso de uso application/imageresize
      → Decoder imaging/resize → Source local à requisição
        → cálculo das dimensões → reamostragem → encoder da família original
```

A aplicação define opções, regras de dimensões, interfaces de decodificação/origem, inspeção e execução. Imaging controla pixels e codecs. Tipos comuns de formato/resultado/erro ficam em `application/imageprocessing`, com aliases em compressão para compatibilidade. Decode/encode HEIF e nome seguro de download são compartilhados por necessidade real das duas funcionalidades.

Inspeção e execução são duas requisições síncronas. O arquivo é enviado e decodificado novamente ao executar. Essa escolha evita introduzir ID, cache persistente de upload, banco, Redis, fila, worker ou object storage apenas para configurar uma imagem. Temporários de multipart/codecs são removidos. O navegador libera URLs Blob na substituição/reset/desmontagem.

O contexto é verificado antes/depois da decodificação e processamento e entre frames. Uma chamada individual de codec não é interrompida à força pelo cancelamento. Isso não representa processamento em background nem isolamento completo de CPU/memória.

## Comportamento e limites

- JPEG, PNG, WebP, AVIF estático, HEIC/HEIF suportado com uma imagem, GIF, BMP e TIFF de uma página mantêm sua família de formato. A compressão existente mantém seu suporte anterior.
- A orientação EXIF de JPEG é normalizada antes do cálculo. AVIF/HEIF seguem o comportamento de orientação dos codecs nativos instalados.
- A reamostragem Catmull–Rom trabalha em RGBA pré-multiplicado. Alpha de PNG/WebP é preservado; resize altera pixels, portanto não é uma operação pixel-identical. BMP mantém as limitações de seu encoder.
- Frames parciais de GIF são compostos respeitando disposal antes da reamostragem. A saída usa frames de canvas completo, paleta web-safe, transparência binária, delays e loop originais. Quantização de cores/paleta e representação interna de disposal podem mudar.
- WebP animado preserva os frames reconstruídos, tempos, loop, fundo e ICC quando presente. O Resize não copia EXIF/XMP para evitar dimensões/orientação desatualizadas. Não há preservação universal de metadados.
- **AVIF animado é rejeitado.** O decoder `gen2brain/avif` v0.6.0 fornece frames e delays, mas não preenche `LoopCount`; por isso não é possível prometer preservação de loops finitos. APNG, TIFF multipágina e HEIF multi-imagem não suportado também são rejeitados, sem achatamento silencioso.
- O multipart completo tem limite de **50 MiB**, com limiar de memória de **8 MiB** no parser. Entrada e saída têm limite de **32 milhões de pixels**. Animações GIF/WebP têm limite de **64 milhões de pixels de canvas-frame**, inclusive para a saída. GIF conta descritores antes de decodificar frames; WebP valida features/quantidade de frames antes da decodificação completa.
- Se as dimensões efetivas forem iguais às originais de exibição, os bytes originais são devolvidos intactos, preservando metadados e evitando recodificação com perda desnecessária.
- Nos demais casos, os parâmetros de encode acompanham os padrões existentes (JPEG/WebP 82, AVIF/HEIF 60; PNG best compression; TIFF Deflate). O resultado pode ser maior em bytes. Não há lote, recorte, conversão, histórico, progresso percentual nem promessa de ganho de detalhe.

Consulte a [documentação de testes](../../TESTS_README.pt-BR.md) para cobertura automatizada e validação da interface.
