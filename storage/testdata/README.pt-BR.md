# Dados de Teste

`storage/testdata/images` contém imagens reais versionadas usadas como fixtures nos testes de integração do backend.

Esses arquivos são usados apenas como entradas para os testes. Eles não são utilizados para armazenar uploads da aplicação, resultados comprimidos, histórico de processamento ou dados de usuários. A aplicação continua processando os uploads de forma síncrona e retorna os bytes da imagem processada diretamente para quem fez a requisição, sem persistir esses arquivos em `storage`.

Os resultados das compressões realizadas durante os testes devem permanecer em memória ou em diretórios temporários dos testes do Go, como `t.TempDir()`. Não grave os arquivos gerados novamente em `storage/testdata/images`.


## Inventário de amostras físicas

As nove amostras versionadas de processamento são estáticas, com 512 × 512 pixels. Os arquivos HEIC/HEIF usam a mesma família.

| Arquivo | Bytes | Família |
| --- | --- | --- |
| [sample.avif](images/sample.avif) | 4351 | avif |
| [sample.bmp](images/sample.bmp) | 1048714 | bmp |
| [sample.gif](images/sample.gif) | 11596 | gif |
| [sample.heic](images/sample.heic) | 6622 | heif |
| [sample.heif](images/sample.heif) | 8391 | heif |
| [sample.jpg](images/sample.jpg) | 30262 | jpeg |
| [sample.png](images/sample.png) | 24487 | png |
| [sample.tiff](images/sample.tiff) | 1048946 | tiff |
| [sample.webp](images/sample.webp) | 7432 | webp |

Compressão e prévia já usam os nove arquivos. Resize verifica quatro transformações por amostra e bytes intactos sem mudança. Conversão verifica nove origens × oito destinos, incluindo nove erros esperados de mesma família. Saídas ficam em memória e a montagem Docker é somente leitura.

Os quatro recursos visuais versionados em `frontend/public/images` (`branding/favicon.ico`, `branding/logo.png`, `hero/grassfield.png`, `hero/mountain.png`) são marca/fundos da interface, não amostras de testes de processamento; também permanecem intactos.
