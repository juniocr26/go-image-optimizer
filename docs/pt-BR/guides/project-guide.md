# Go Image Optimizer

[English](../../en/guides/project-guide.md) | [Português](project-guide.md)

Aplicação para otimização de imagens desenvolvida em Go, com uma interface web utilizando Next.js, React e Tailwind CSS.

Compressão, Resize e conversão são síncronos e devolvem as imagens processadas diretamente ao navegador. Prévias nativas contam com fallback gerado pelo servidor.

> **Status atual:** Completo para o escopo funcional atual: compressão, redimensionamento, conversão de formatos e fallback de prévia compatível com navegadores. Não há funcionalidades adicionais planejadas neste momento. Manutenção e correções continuam possíveis; isso não significa prontidão para produção.

## Visão geral

O Go Image Optimizer é um projeto de portfólio voltado ao processamento de imagens e à aplicação de conceitos de engenharia de backend utilizando Go.

Em vez de definir uma arquitetura complexa antecipadamente, o projeto segue uma abordagem incremental: começar com uma solução simples, validar os requisitos e introduzir mudanças arquiteturais quando existir uma razão concreta para isso.

## Escopo atual

A aplicação permite:

- Enviar uma imagem pela interface web.
- Enviar essa imagem para o backend em Go.
- Comprimir, configurar Resize por pixels/porcentagem ou converter o formato em um modal.
- Visualizar imagens suportadas pelo navegador ou pelo fallback apenas para exibição.
- Receber a imagem processada com informações reais do resultado.
- Baixar o resultado diretamente pelo navegador.

O backend identifica o formato real da imagem pelos bytes do arquivo, sem confiar em extensão ou MIME type informado pelo navegador. Compressão e Resize mantêm a família de origem. A compressão preserva dimensões orientadas, Resize as altera explicitamente e a conversão preserva dimensões orientadas ao mudar a família. Nenhuma operação garante um arquivo menor.

O escopo é deliberadamente focado em um portfólio Go e Next.js. Resize atende à necessidade básica de imagens menores. Operação separada de thumbnails, histórico, bancos de dados, Redis, filas, workers, persistência de imagens/object storage e editor semelhante ao Paint estão fora do escopo. O fallback de prévia é um recurso interno de exibição, não outra operação do usuário.

## Formatos de compressão

| Formato     | Saída               | Observações                                                                                                                           |
| ----------- | ------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| JPEG / JPG  | JPEG                | Re-encode lossy com qualidade conservadora. A orientação EXIF é aplicada aos pixels antes da saída.                                   |
| PNG         | PNG                 | Saída lossless para pixels usando compressão PNG alta. Transparência é preservada.                                                    |
| WebP        | WebP                | WebP estático e animado são suportados. A saída animada preserva quantidade de frames e tempos, reencodando frames reconstruídos.     |
| AVIF        | AVIF                | AVIF estático é suportado. A implementação usa a API multi-imagem de AVIF, mas a cobertura automatizada atual usa fixtures estáticas. |
| HEIC / HEIF | Família HEIC / HEIF | Usa libheif/HEVC nativo no container do backend. Variantes HEIF não suportadas são rejeitadas em vez de simuladas.                    |
| GIF         | GIF                 | GIF estático e animado são suportados, incluindo delays e configuração de loop.                                                       |
| BMP         | BMP                 | Decodificado e reencodado como BMP; redução de tamanho não é garantida.                                                               |
| TIFF        | TIFF                | Reencodado como TIFF com compressão Deflate.                                                                                          |

WebM, SVG, formatos RAW de câmera, vídeos e arquivos compactados não são suportados.

## Tecnologias

### Backend

- Go
- Bibliotecas nativas libheif no runtime Docker para suporte a HEIC/HEIF

### Frontend

- Next.js
- React
- Tailwind CSS

## Conclusão do escopo

Compressão, redimensionamento, conversão e fallback de prévia estão implementados. Não há roadmap de novas funcionalidades além desse escopo. Compatibilidade, dependências nativas e limites de recursos/cancelamento continuam sendo limitações técnicas documentadas. Uma evolução hipotética para produção exigiria novos requisitos e medições, sem infraestrutura adicional apenas para aumentar a complexidade do portfólio.

## Licença

Este projeto é licenciado sob a MIT License. Consulte o arquivo `LICENSE` para mais detalhes.

## Conversão de formatos

Escolha arquivo estático suportado, inspecione a família detectada pelos bytes e selecione destino diferente. Conversão preserva dimensões orientadas, com políticas de alpha/metadados por formato e tamanhos reais; resultado maior continua disponível. Animações e variantes multi-imagem não suportadas são rejeitadas. A [arquitetura](../architecture/overview.md#conversão-de-formatos) mantém contrato de campos/headers, defaults e limites.

## Prévias de imagens

Tenta exibição nativa primeiro. Se falhar, thumbnail estático somente para display ajuda sem trocar bytes do original/download. Erro/retry ficam na área de prévia. Cache de sessão compartilha requisições pela identidade do Blob; reload perde estado temporário. Veja [arquitetura de prévias](../architecture/overview.md#prévias-de-imagens) para variantes, recursos e ciclo de vida.

[Testes](../testing/strategy.md) · [Docker](../docker/runtime.md) · [Desenvolvimento e recuperação](../docker/development.md) · [Verificação](../testing/verification.md)
