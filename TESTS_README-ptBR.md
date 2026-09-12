# Documentação de Testes

Este documento descreve a estratégia atual de testes do Go Image Optimizer.

## Estratégia

O backend possui testes automatizados em Go para a fronteira de aplicação, a implementação de compressão de imagens e o contrato HTTP. O frontend, neste MVP, é validado pelo build de produção e por validação manual do fluxo; nenhum framework de testes de frontend foi adicionado.

Os testes evitam afirmar que toda imagem otimizada precisa ficar menor. Algumas imagens reais já chegam otimizadas. As asserções de redução de tamanho ficam limitadas a fixtures determinísticas criadas especificamente para esse caso.

## Cobertura do backend

Os testes automatizados do backend cobrem:

- comportamento do caso de uso de compressão;
- tratamento de contexto cancelado;
- compressão válida de JPEG/JPG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP e TIFF;
- imagens de saída podem ser decodificadas;
- dimensões são preservadas;
- formato da resposta é preservado;
- conteúdo de pixels do PNG permanece lossless;
- transparência de WebP lossless permanece lossless;
- quantidade de frames, delays e loop de GIF animado;
- quantidade de frames e duração dos frames de WebP animado;
- normalização de orientação EXIF em JPEG;
- entrada não suportada é rejeitada;
- imagem corrompida é rejeitada para cada assinatura de formato suportado;
- limite de segurança por quantidade de pixels decodificados para todos os formatos suportados;
- limite de segurança para pixels de canvas-frame em animações;
- validação de campo de imagem ausente;
- requisições multipart malformadas;
- requisições que não são multipart;
- proteção do limite de 25 MiB da requisição;
- headers de resposta em sucesso;
- nomes de download gerados;
- spoofing de extensão, em que o tipo de resposta segue os bytes detectados e não o nome enviado.

## Validação do frontend

O build do frontend valida TypeScript e a compilação de produção.

A validação manual da interface deve cobrir:

- seleção por drag and drop;
- seleção pelo file picker;
- previews para formatos que o navegador consegue renderizar;
- placeholder para formatos que o navegador não consegue pré-visualizar, como muitos arquivos HEIC ou TIFF;
- comportamento do botão Compress;
- bloqueio de envios duplicados durante a compressão;
- estado de carregamento indeterminado;
- preview do resultado;
- exibição de tamanho original, tamanho otimizado e redução usando bytes reais;
- tratamento honesto quando o resultado otimizado não fica menor;
- convenção de nome do arquivo baixado;
- reset / processamento de outra imagem;
- limpeza de URLs Blob em trocas de seleção, reset e desmontagem do componente.

## Como executar

Testes do backend com Go instalado localmente:

```bash
cd backend
go test ./...
```

Testes locais de HEIC/HEIF exigem bibliotecas de desenvolvimento nativas da libheif e plugins de codec HEVC. O Docker é o caminho recomendado quando essas dependências não estão instaladas localmente.

Testes do backend com Docker:

```bash
docker run --rm -v "$PWD/backend:/src" -w /src golang:1.27.1-alpine sh -lc \
  'apk add --no-cache build-base pkgconf libheif-dev libheif-libde265 libheif-x265 >/dev/null && /usr/local/go/bin/go test ./...'
```

Build de produção do frontend:

```bash
cd frontend
npm run build
```

Validação smoke com Docker:

```bash
docker compose up --build
```

Depois, abra o frontend, envie amostras representativas de JPEG/JPG, PNG, WebP, AVIF, HEIC/HEIF, GIF, BMP e TIFF, comprima as imagens, baixe os resultados e confirme que a configuração do Compose não monta um volume de armazenamento da aplicação para resultados de imagem.

## Limitações atuais

- Ainda não há testes automatizados de interação no navegador.
- Não há asserções visuais de qualidade para a saída JPEG.
- O fluxo do frontend é validado manualmente.
- A validação com Docker Compose é um smoke test, não um teste de carga ou escalabilidade.
- WebM, SVG, RAW, vídeo e arquivos compactados são intencionalmente não suportados e entram na cobertura como comportamento de entrada não suportada, não como testes de codec.
- Nenhum percentual de cobertura é declarado.
