# Dados de Teste

`storage/testdata/images` contém imagens reais versionadas usadas como fixtures nos testes de integração do backend.

Esses arquivos são usados apenas como entradas para os testes. Eles não são utilizados para armazenar uploads da aplicação, resultados comprimidos, histórico de processamento ou dados de usuários. A aplicação continua processando os uploads de forma síncrona e retorna os bytes da imagem comprimida diretamente para quem fez a requisição, sem persistir esses arquivos em `storage`.

Os resultados das compressões realizadas durante os testes devem permanecer em memória ou em diretórios temporários dos testes do Go, como `t.TempDir()`. Não grave os arquivos gerados novamente em `storage/testdata/images`.
