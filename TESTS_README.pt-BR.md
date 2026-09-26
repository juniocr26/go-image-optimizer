# Documentação de Testes

Este documento descreve a estratégia atual de testes do Go Image Optimizer.

## Estratégia

O backend possui testes automatizados em Go para a fronteiras de aplicação, as implementações de compressão e resize, os contratos HTTP, regressões determinísticas de codec e fixtures reais de integração. O frontend tem testes de regressão do cálculo de dimensões usando o runner nativo do Node.js, verificações TypeScript/build e validação manual do fluxo; não há framework de testes de navegador instalado.

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
docker compose run --rm --no-deps frontend-dev node --test tests/resize-options.test.mjs
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
- `go test -race ./...` está bloqueado no momento por uma falha de `checkptr` dentro de `github.com/strukturag/libheif` durante a geração das fixtures HEIC/HEIF; a suíte normal sem `-race` passa.
- WebM, SVG, RAW, vídeo e arquivos compactados são intencionalmente não suportados e entram na cobertura como comportamento de entrada não suportada, não como testes de codec.
- Nenhum percentual de cobertura é declarado.

## Cobertura de Resize

- Testes de cálculo no frontend cobrem dimensões inalteradas, ampliação por ambas as âncoras, dimensões independentes, três reduções, arredondamento, entradas inválidas e limites de pixels/frames de saída. Usam o Node.js 24 da imagem Docker do frontend.

- Testes de aplicação cobrem os dois modos, arredondamento de dimensões ímpares, mínimo de um pixel, âncora de largura/altura, distorção, ampliação proporcional por ambos os eixos e ampliação independente, parâmetros inválidos, limites de saída/animação, proporções extremas, cancelamento e retorno dos bytes originais sem alteração de dimensões.
- Testes de imaging redimensionam as nove fixtures reais, decodificam a saída independentemente e verificam dimensões, formato detectado e MIME. Requisições sem alteração de dimensões são verificadas byte a byte nas nove fixtures. Testes sintéticos cobrem EXIF JPEG e posição dos pixels após rotação, alpha PNG/WebP lossless, GIF com frames parciais e disposal previous/background, delays/loops GIF, tempos/loops WebP, input inválido, limites de pixels PNG antes do decode e limites de frames GIF/WebP. AVIF animado é rejeitado explicitamente porque o decoder instalado perde metadados de loop.
- Integração HTTP cobre inspeção, headers, spoofing de formato/nome, ampliação padrão por ambas as âncoras, dimensões independentes, bytes originais quando o tamanho não muda, dimensões decodificadas correspondentes aos headers, opções inválidas (fracionários/overflow/NaN), multipart malformado/ambíguo, opções duplicadas, booleanos vazios, status de variante não suportada, limite de upload e limite de pixels de saída.
- `docker compose run --rm backend-test` inclui Resize e regressões de compressão. Os codecs HEIF nativos continuam necessários. A limitação de `-race` acima continua aplicável.

Validação de interface: Run abre a configuração sem executar Resize; dimensões consideram orientação; editar um campo atualiza o outro; destravar permite distorção; ampliação mostra a nota sobre detalhes e corresponde às dimensões baixadas; porcentagens indicam redução linear; valores inválidos impedem envio; loading/erros ficam no modal; retry funciona; dimensões/download são reais; cancelar/Escape restauram foco; Tab permanece no dialog; conteúdo mobile rola com rodapé acessível; fallback HEIC/TIFF permite processar; repetir parte do original. Revalidar compressão após Resize. São verificações manuais do fluxo no navegador, separadas dos testes automatizados de cálculo.

Consulte [comportamento/API do Resize](docs/pt-BR/architecture.md#redimensionamento-de-imagens).
