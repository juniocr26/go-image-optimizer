# Documentação de Testes

[English](../en/testing.md) | [Português](testing.md)

Este documento descreve a estratégia atual de testes do Go Image Optimizer.

## Estratégia

O backend possui testes automatizados em Go para as fronteiras de aplicação, as implementações de compressão, resize e conversão, os contratos HTTP, regressões determinísticas de codec e fixtures reais de integração. O frontend tem testes de cálculo de dimensões, apresentação de tamanhos da conversão e cache de prévia usando o runner nativo do Node.js, verificações TypeScript/build e checklist manual do fluxo; não há framework de testes de navegador instalado.

Os testes evitam afirmar que toda imagem otimizada precisa ficar menor. Algumas imagens reais já chegam otimizadas. As asserções de redução de tamanho ficam limitadas a fixtures determinísticas criadas especificamente para esse caso.

## Cobertura do backend

Os testes automatizados do backend cobrem:

- comportamento do caso de uso de compressão;
- tratamento de contexto cancelado;
- compressão válida de JPEG/JPG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP e TIFF;
- compressão de fixtures reais em `storage/testdata/images`;
- detecção dos bytes de saída comprimidos para todos os formatos estáticos suportados;
- imagens de saída podem ser decodificadas;
- dimensões são preservadas;
- formato da resposta é preservado;
- conteúdo de pixels do PNG permanece lossless;
- transparência de WebP lossless permanece lossless;
- quantidade de frames, delays e loop de GIF animado;
- quantidade de frames e duração dos frames de WebP animado;
- normalização de orientação EXIF em JPEG;
- entrada não suportada é rejeitada, incluindo assinaturas de RAW de câmera;
- imagem corrompida é rejeitada para cada assinatura de formato suportado;
- limite de segurança por quantidade de pixels decodificados para todos os formatos suportados;
- limite de segurança para pixels de canvas-frame em animações;
- validação de campo de imagem ausente;
- requisições multipart malformadas;
- requisições que não são multipart;
- proteção do limite de 50 MiB da requisição;
- headers de resposta em sucesso;
- nomes de download gerados;
- spoofing de extensão, em que o tipo de resposta segue os bytes detectados e não o nome enviado.

## Matriz de cobertura de formatos da compressão

| Formato | Compressão invocada | Saída decodificada | Formato preservado | Dimensões verificadas | Propriedades específicas verificadas |
| --- | --- | --- | --- | --- | --- |
| JPEG / JPG | Compressor real com bytes sintéticos e `sample.jpg`; rota HTTP com bytes JPEG sintéticos | `jpeg.Decode` | `FormatJPEG`, `image/jpeg`, bytes de saída detectados, nomes `.jpg` / `.jpeg` | Sim | Normalização de orientação EXIF e fixture determinística de redução de JPEG em alta qualidade |
| PNG | Compressor real com bytes sintéticos e `sample.png`; rota HTTP com bytes PNG sintéticos | `png.Decode` | `FormatPNG`, `image/png`, bytes de saída detectados | Sim | Preservação lossless de pixels e alpha na fixture sintética; igualdade de pixels da fixture real após re-encode |
| WebP | Compressor real com bytes sintéticos e `sample.webp`; caminho sintético animado também exercitado | `webp.Decode`; `animation.Decode` quando animado | `FormatWebP`, `image/webp`, bytes de saída detectados | Sim | Transparência lossless na fixture sintética; quantidade de frames animados e duração dos frames |
| AVIF | Compressor real com bytes sintéticos e `sample.avif` | `avif.Decode`; `avif.DecodeAll` para metadados de frames | `FormatAVIF`, `image/avif`, bytes de saída detectados | Sim | Fixture real AVIF estática; comportamento multi-frame não é declarado além do roteamento suportado pelo codec |
| HEIC | Compressor real com bytes sintéticos via libheif e `sample.heic` | Decode da imagem primária via libheif | `FormatHEIF`, `image/heic`, bytes de saída detectados | Sim | Caminho nativo de encode/decode via libheif/HEVC é exercitado |
| HEIF | Compressor real com `sample.heif` | Decode da imagem primária via libheif | `FormatHEIF`, saída detectada como família HEIC/HEIF | Sim | Arquivo versionado `.heif` separado é aceito pelo caminho HEIF nativo |
| GIF | Compressor real com bytes sintéticos e `sample.gif`; caminho sintético animado também exercitado | `gif.Decode`; `gif.DecodeAll` para metadados de animação | `FormatGIF`, `image/gif`, bytes de saída detectados | Sim | Quantidade de frames, delays, loop e valores de disposal na fixture sintética animada; metadados da fixture real continuam válidos |
| BMP | Compressor real com bytes sintéticos e `sample.bmp` | `bmp.Decode` | `FormatBMP`, `image/bmp`, bytes de saída detectados | Sim | Re-encode BMP bem-sucedido sem exigir redução de tamanho |
| TIFF | Compressor real com bytes sintéticos e `sample.tiff`; rota HTTP cobre nomes `.tif` / `.tiff` | `tiff.Decode` | `FormatTIFF`, `image/tiff`, bytes de saída detectados | Sim | Re-encode TIFF com compressão Deflate; spoofing RAW/DNG continua rejeitado |

## Estratégia das imagens de teste

A suíte normal de testes não baixa imagens em tempo de execução.

As imagens sintéticas de teste são determinísticas e geradas por helpers nos testes Go:

- gradientes e padrões detalhados de pixels para cobertura de codecs estáticos;
- padrões com alpha/transparência para PNG e WebP lossless;
- pequenas amostras animadas de GIF e WebP para comportamento de frames e tempos;
- um payload PNG grande e determinístico para comportamento do limite de upload;
- bytes HEIC/HEIF de origem gerados via libheif e HEVC em um diretório temporário do teste.

As fixtures reais de integração são versionadas em `storage/testdata/images`:

- `sample.jpg`
- `sample.png`
- `sample.webp`
- `sample.avif`
- `sample.heic`
- `sample.heif`
- `sample.gif`
- `sample.bmp`
- `sample.tiff`

Essas fixtures são entradas de teste commitadas, não armazenamento da aplicação. Os testes com arquivos reais leem esses arquivos, comprimem os bytes, validam a resposta e descartam a saída comprimida em memória. Os testes não devem gravar arquivos gerados de volta em `storage/testdata/images`.

Os testes de HEIC/HEIF exigem bibliotecas de desenvolvimento nativas da libheif e plugins HEVC de decode/encode porque tanto a geração sintética de HEIC quanto a compressão real de HEIC/HEIF usam o caminho do codec nativo.

## Validação do frontend

O build do frontend valida TypeScript e a compilação de produção.

A validação manual da interface deve cobrir:

- seleção por drag and drop;
- seleção pelo file picker;
- previews para formatos que o navegador consegue renderizar;
- placeholder para formatos que o navegador não consegue pré-visualizar, como muitos arquivos HEIC ou TIFF;
- seleção da operação e comportamento do botão Run;
- bloqueio de envios duplicados durante a compressão;
- estado de carregamento indeterminado;
- preview do resultado;
- exibição de tamanho original, tamanho otimizado e redução usando bytes reais;
- tratamento honesto quando o resultado otimizado não fica menor;
- convenção de nome do arquivo baixado;
- reset / processamento de outra imagem;
- limpeza de URLs Blob em trocas de seleção, reset e desmontagem do componente.

## Como executar

Testes recomendados do backend com Docker Compose:

```bash
docker compose run --rm backend-test
```

Esse comando usa o serviço Compose `backend-test`, que aponta para o estágio de build Go do `backend/Dockerfile`. Essa imagem contém o toolchain Go, ferramentas de build com CGO, `pkgconf`, headers de desenvolvimento da libheif, `libheif-libde265` e `libheif-x265`. O serviço monta `./backend` em `/src` e monta `./storage/testdata/images` como somente leitura em `/testdata/images`, então alterações no código-fonte e fixtures reais versionadas ficam visíveis sem rebuildar a imagem de runtime de produção.

O serviço `backend-test` fica atrás do profile `test` e não é iniciado pelo `docker compose up` normal. O container da API backend não precisa estar rodando para o comando de teste acima. Se o projeto já estiver rodando, execute o comando de teste em outro terminal.

Construa a imagem de teste uma vez durante o setup inicial ou depois de alterar `backend/Dockerfile`, pacotes nativos, versão do Go, `go.mod` ou `go.sum`:

```bash
docker compose build backend-test
```

O comando padrão executa `go test -v ./...`, então nomes de pacotes, testes, subtestes e formatos ficam visíveis. Linhas como `? github.com/.../cmd/api [no test files]` são saída normal do Go para pacotes que intencionalmente não têm arquivos de teste; elas não representam falha.

Para passar flags customizadas do Go test no mesmo ambiente:

```bash
docker compose run --rm backend-test go test -v ./internal/infrastructure/imaging/compress -run TestCompressorRealFixtures
```

Testes do backend com Go instalado localmente continuam suportados quando as dependências nativas correspondentes estão instaladas localmente:

```bash
cd backend
CGO_ENABLED=1 go test -v ./...
```

Testes locais de HEIC/HEIF exigem bibliotecas de desenvolvimento nativas da libheif e plugins de codec HEVC. O Docker Compose é o caminho recomendado quando essas dependências não estão instaladas localmente.

Verificação TypeScript e build de produção no ambiente Docker documentado:

```bash
docker compose run --rm --no-deps frontend-dev node --test tests/resize-options.test.mjs tests/conversion-size.test.mjs tests/preview-cache.test.mjs
docker compose run --rm --no-deps frontend-dev npx tsc --noEmit
docker compose run --rm --no-deps -e NODE_ENV=production frontend-dev npm run build
```

A variável de produção é necessária porque `frontend-dev` define `NODE_ENV=development`. Não há script de lint no frontend.

Validação smoke com Docker:

```bash
docker compose up --build
```

Depois, abra o frontend, envie amostras representativas de JPEG/JPG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP e TIFF, comprima as imagens, baixe os resultados e confirme que a configuração do Compose não monta um volume de armazenamento da aplicação para resultados de imagem.

## Limitações atuais

- Não há suíte de interação no navegador versionada.
- Não há asserções visuais de qualidade para a saída JPEG.
- TypeScript/build não verificam interações no navegador nem certificam compatibilidade entre navegadores.
- A validação com Docker Compose é um smoke test, não um teste de carga ou escalabilidade.
- Uma execução anteriormente documentada de `go test -race ./...` encontrou falha de `checkptr` dentro de `github.com/strukturag/libheif` durante a geração das fixtures HEIC/HEIF. O race detector não foi reexecutado nesta tarefa; o fluxo recomendado usa a suíte normal.
- WebM, SVG, RAW, vídeo e arquivos compactados são intencionalmente não suportados e entram na cobertura como comportamento de entrada não suportada, não como testes de codec.
- Nenhum percentual de cobertura é declarado.

## Cobertura de Resize

- Testes de cálculo no frontend cobrem dimensões inalteradas, ampliação por ambas as âncoras, dimensões independentes, três reduções, arredondamento, entradas inválidas e limites de pixels/frames de saída. Usam o Node.js 24 da imagem Docker do frontend.

- Testes de aplicação cobrem os dois modos, arredondamento de dimensões ímpares, mínimo de um pixel, âncora de largura/altura, distorção, ampliação proporcional por ambos os eixos e ampliação independente, parâmetros inválidos, limites de saída/animação, proporções extremas, cancelamento e retorno dos bytes originais sem alteração de dimensões.
- Testes de imaging redimensionam as nove fixtures reais, decodificam a saída independentemente e verificam dimensões, formato detectado e MIME. Requisições sem alteração de dimensões são verificadas byte a byte nas nove fixtures. Testes sintéticos cobrem EXIF JPEG e posição dos pixels após rotação, alpha PNG/WebP lossless, GIF com frames parciais e disposal previous/background, delays/loops GIF, tempos/loops WebP, input inválido, limites de pixels PNG antes do decode e limites de frames GIF/WebP. AVIF animado é rejeitado explicitamente porque o decoder instalado perde metadados de loop.
- Integração HTTP cobre inspeção, headers, spoofing de formato/nome, ampliação padrão por ambas as âncoras, dimensões independentes, bytes originais quando o tamanho não muda, dimensões decodificadas correspondentes aos headers, opções inválidas (fracionários/overflow/NaN), multipart malformado/ambíguo, opções duplicadas, booleanos vazios, status de variante não suportada, limite de upload e limite de pixels de saída.
- `docker compose run --rm backend-test` inclui Resize e regressões de compressão. Os codecs HEIF nativos continuam necessários. A limitação de `-race` acima continua aplicável.

Validação de interface: Run abre a configuração sem executar Resize; dimensões consideram orientação; editar um campo atualiza o outro; destravar permite distorção; ampliação mostra a nota sobre detalhes e corresponde às dimensões baixadas; porcentagens indicam redução linear; valores inválidos impedem envio; loading/erros ficam no modal; retry funciona; dimensões/download são reais; cancelar/Escape restauram foco; Tab permanece no dialog; conteúdo mobile rola com rodapé acessível; fallback HEIC/TIFF permite processar; repetir parte do original. Revalidar compressão após Resize. São verificações manuais do fluxo no navegador, separadas dos testes automatizados de cálculo.

Consulte [comportamento/API do Resize](architecture.md#redimensionamento-de-imagens).

## Validação da conversão

```sh
docker compose run --rm backend-test
docker compose run --rm --no-deps frontend-dev node --test tests/resize-options.test.mjs tests/conversion-size.test.mjs tests/preview-cache.test.mjs
docker compose run --rm --no-deps frontend-dev npx tsc --noEmit
docker compose run --rm --no-deps -e NODE_ENV=production frontend-dev npm run build
```

Os testes cobrem famílias de saída, decodificação, alfa/fundo branco, orientação EXIF, rejeição de animação, cancelamento, contrato HTTP, duplicatas e saída maior. Os testes existentes continuam cobrindo limites e variantes do decodificador compartilhado e regressões de compressão/resize. A verificação manual no navegador deve cobrir desktop/mobile, Tab/Shift+Tab/Escape, loading/retry, erro com destino preservado, fallback de prévia e download; não foi executada durante a conclusão do escopo.

Os testes de prévia cobrem fixtures reais, MIME e decodificação independente, limites da miniatura, alpha, orientação, entradas inválidas, limites de pixels/upload e variantes, validação multipart, preservação da origem, envio do Blob de resultado, cache/retry e cancelamento. Execute `node --test tests/*.test.mjs` no container frontend-dev documentado. A verificação de interação no navegador não foi realizada.


## Matriz de conversão com amostras reais

`imaging/convert/real_fixtures_test.go` testa cada um dos nove arquivos físicos contra oito destinos canônicos (`jpeg`, `png`, `webp`, `avif`, `heif`, `gif`, `bmp`, `tiff`). São 72 casos nomeados: 63 conversões válidas e nove erros esperados de mesma família. JPG é alias de extensão/interface; `targetFormat=jpg` é rejeitado. HEIC e HEIF representam a família `heif` e rejeitam esse destino. Alterar o inventário sem definir uma política faz o teste falhar.

As nove amostras físicas atuais são estáticas, com 512 × 512 pixels. Decode direto pelos codecs verifica bytes e dimensões orientadas independentemente do Decoder de processamento. Os testes verificam família detectada, MIME, informações de origem/resultado, saída de um frame e bytes de origem intactos. Pixels totalmente transparentes, quando presentes, exercitam preservação de alfa ou composição branca; fixtures sintéticas determinísticas também verificam transparência, alfa parcial e orientação EXIF. Nenhuma conversão exige redução universal de bytes.

Os testes sintéticos verificam rejeição de animações, APNG, sequências AVIF pelo decoder compartilhado de Resize e TIFF multipágina. A matriz real não prova suporte a variantes arbitrárias. Testes HTTP usam imagens sintéticas pequenas para verificar MIME, extensões, nomes seguros, dimensões, no-store e Content-Length igual ao corpo para todos os destinos, sem repetir a matriz cara. Testes de frontend verificam redução, aumento e ausência de mudança de tamanho.

## Resize com amostras reais

O teste existente foi ampliado, sem duplicação: cada uma das nove amostras passa por redução/ampliação proporcionais e redução/ampliação com proporção destravada, totalizando 36 transformações. Origem e resultado são decodificados para conferir dimensões, família, MIME e propriedades relevantes de alfa/frames. Como todas são quadradas, a redução independente usa saída 256 × 128 e a ampliação 515 × 519; a ampliação proporcional usa 516 × 516. A redução proporcional produz 256 × 256. Os nove casos sem mudança continuam verificando igualdade dos bytes. Orientação, transparência parcial, tempos/loops e variantes rejeitadas mantêm testes sintéticos próprios.

## Verificação da conclusão do escopo em 04 de outubro de 2026

Executados com sucesso no ambiente Docker documentado:

```sh
docker compose run --rm backend-test
docker compose run --rm backend-test sh -c 'gofmt -w internal/infrastructure/imaging/resize/resize_test.go internal/infrastructure/imaging/convert/real_fixtures_test.go && go test -v ./internal/infrastructure/imaging/resize ./internal/infrastructure/imaging/convert'
docker compose run --rm backend-test sh -c 'gofmt -w internal/infrastructure/http/conversion_test.go internal/infrastructure/imaging/resize/resize_test.go && go test ./...'
docker compose run --rm backend-test sh -c 'gofmt -w internal/infrastructure/imaging/convert/processor_test.go && go test -count=1 ./...'
docker compose run --rm --no-deps frontend-dev node --test tests/resize-options.test.mjs tests/conversion-size.test.mjs tests/preview-cache.test.mjs
docker compose run --rm --no-deps frontend-dev npx tsc --noEmit
docker compose run --rm --no-deps -e NODE_ENV=production frontend-dev npm run build
```

A primeira execução estabeleceu a base; a suíte completa final inclui amostras ampliadas e headers de todos os destinos. Os 12 testes de frontend passaram. São verificações automatizadas neste ambiente, não interação de navegador, fidelidade visual, segurança com race detector, benchmarks ou capacidade de produção. Não foram executados testes de navegador, race ou carga. A falha de checkptr/libheif descrita anteriormente é um registro anterior, não uma nova execução desta tarefa. Resultados ficaram em memória; arquivos temporários HEIF são removidos pelo helper compartilhado. Os bytes originais das fixtures foram comparados ao Git e permaneceram intactos.

O guia DOCX foi validado estruturalmente (XML, estilos de título, 32 respostas, duas tabelas e campo de página). A tentativa de renderização pelo script da skill falhou por falta de `pdf2image`; o runtime empacotado de documentos/LibreOffice não está disponível nesta sessão. A paginação e o layout visual do guia permanecem sem inspeção.

## Revisão documental de arquitetura — 2026-10-05

O build stage Docker existente foi construído com tag temporária `portfolio-doc-review-go-tests`. Container descartável executou `go test -count=1 ./...`, CGO habilitado, backend atual somente leitura, fixtures reais somente leitura em `/testdata/images` e rede desabilitada. Todos os pacotes com testes passaram, incluindo HEIF nativo e fixtures reais. Imagem runtime frontend existente executou os 12 testes de cálculo/cache sobre fonte atual somente leitura; todos passaram. A primeira chamada frontend tinha argumento `node` duplicado e falhou antes dos testes; chamada corrigida passou. Sem stack da aplicação, interação browser, build/TypeScript frontend, race ou carga. Containers e tag temporária removidos; fixtures existentes preservadas. Caches de build podem permanecer no Docker.


## Inventário de fixtures

`storage/testdata/images` contém imagens reais versionadas usadas como fixtures nos testes de integração do backend.

Esses arquivos são usados apenas como entradas para os testes. Eles não são utilizados para armazenar uploads da aplicação, resultados comprimidos, histórico de processamento ou dados de usuários. A aplicação continua processando os uploads de forma síncrona e retorna os bytes da imagem processada diretamente para quem fez a requisição, sem persistir esses arquivos em `storage`.

Os resultados das compressões realizadas durante os testes devem permanecer em memória ou em diretórios temporários dos testes do Go, como `t.TempDir()`. Não grave os arquivos gerados novamente em `storage/testdata/images`.


## Inventário de amostras físicas

As nove amostras versionadas de processamento são estáticas, com 512 × 512 pixels. Os arquivos HEIC/HEIF usam a mesma família.

| Arquivo | Bytes | Família |
| --- | --- | --- |
| [sample.avif](../../storage/testdata/images/sample.avif) | 4351 | avif |
| [sample.bmp](../../storage/testdata/images/sample.bmp) | 1048714 | bmp |
| [sample.gif](../../storage/testdata/images/sample.gif) | 11596 | gif |
| [sample.heic](../../storage/testdata/images/sample.heic) | 6622 | heif |
| [sample.heif](../../storage/testdata/images/sample.heif) | 8391 | heif |
| [sample.jpg](../../storage/testdata/images/sample.jpg) | 30262 | jpeg |
| [sample.png](../../storage/testdata/images/sample.png) | 24487 | png |
| [sample.tiff](../../storage/testdata/images/sample.tiff) | 1048946 | tiff |
| [sample.webp](../../storage/testdata/images/sample.webp) | 7432 | webp |

Compressão e prévia já usam os nove arquivos. Resize verifica quatro transformações por amostra e bytes intactos sem mudança. Conversão verifica nove origens × oito destinos, incluindo nove erros esperados de mesma família. Saídas ficam em memória e a montagem Docker é somente leitura.

Os quatro recursos visuais versionados em `frontend/public/images` (`branding/favicon.ico`, `branding/logo.png`, `hero/grassfield.png`, `hero/mountain.png`) são marca/fundos da interface, não amostras de testes de processamento; também permanecem intactos.
