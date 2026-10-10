# Limites de recursos e código nativo

[English](../../en/security/resource-boundaries.md) | [Português brasileiro](resource-boundaries.md)

Revisão estática do código: 2026-10-10. Fatos implementados, teoria geral e mudanças hipotéticas são separados abaixo. Comandos runtime não foram executados.

Bytes codificados não equivalem à memória decodificada. A camada de imagens limita dimensões estáticas a 32.000.000 pixels e trabalho de canvas animado a 64.000.000 pixels de frames. A validação usa divisão para evitar overflow de multiplicação. Um buffer RGBA no limite estático já representa cerca de 128 milhões de bytes; original, destino, frames, buffers do proxy e alocações do codec se somam. Limites são por operação, com verificações por variante; não são cota global de memória ou limite de concorrência. Parte da decodificação nativa pode preceder verificações posteriores.

`UseCase.Execute` verifica contexto antes e depois da compressão. Não passa um contexto cancelável ao compressor. Propagar abortamento e definir timeouts HTTP, portanto, não prova interrupção de um codec nativo síncrono. O servidor Go configura 5 s para cabeçalhos, 2 min para leitura/gravação e 60 s de ociosidade. Prazos HTTP limitam transporte; o encoder pode continuar consumindo CPU até retornar. Um semáforo ou processo de worker separado poderia oferecer admissão e isolamento, com custo de complexidade e comunicação; nenhum existe.

CGO e codecs nativos ampliam suporte HEIF/AVIF e acrescentam manutenção de ABI, build e segurança. Temporários multipart são removidos após interpretação bem-sucedida; handles são fechados; helpers HEIF limpam seus temporários. Resultados por requisição não são armazenamento durável. Recuperação de panic no handler não isola biblioteca nativa nem recupera toda falha nativa do processo. Não há login, terminação TLS ou rate limit por usuário implementados na aplicação revisada. Exposição pública exige projeto separado; validação do conteúdo sozinha não torna seguros uploads hostis.
