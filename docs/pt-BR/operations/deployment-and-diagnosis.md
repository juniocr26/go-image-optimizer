# Implantação, diagnóstico e evidência

[English](../../en/operations/deployment-and-diagnosis.md) | [Português brasileiro](deployment-and-diagnosis.md)

Revisão estática do código: 2026-10-10. Fatos implementados, teoria geral e mudanças hipotéticas são separados abaixo. Comandos runtime não foram executados.

Compose descreve backend Go e frontend Next.js, além de perfis de teste/desenvolvimento. O runtime do backend exige bibliotecas nativas fornecidas pelo Dockerfile. Bind mounts e caches de dependências de desenvolvimento ajudam a edição; não são procedimento de rollout em produção. `BACKEND_URL` do servidor frontend precisa resolver a partir do próprio runtime, geralmente `backend:8080` no Compose. `localhost` ali significa aquele contêiner, não o host do desenvolvedor ou outro serviço.

Diagnostique separadamente: 502 do proxy sugere falha de encaminhamento/abortamento/conectividade; 4xx do backend identifica entrada inválida ou não suportada; 500 de codec sugere falha de implantação ou encoder; `/health` é liveness, não autoteste de todos os codecs. Analise warnings/erros estruturados de `slog` sem registrar conteúdo da imagem. Preserve um reproducer mínimo não sensível e seu formato/variante real. Testes HTTP, fakes de aplicação e fixtures de imagens comprovam fronteiras diferentes; builds/testes históricos registrados não são nova validação. Nenhum navegador, carga, teste de corrida ou aplicação foi executado nesta revisão.

Não há banco persistente de uploads para backup nem migrações. Atualizar o backend exige manter dependências de codec compatíveis e os metadados de resposta do frontend/backend precisam continuar compatíveis. Rollback de imagens é conceitualmente possível com artefatos retidos, mas não há evidência de registro de releases, pipeline produtivo, topologia TLS, ensaio de rollback ou SLO de produção. Benchmarks não se aplicam sem medições: limites de pixels e bytes mostrados na UI não medem vazão. Lacunas incluem picos de memória concorrente, interrupção de trabalho nativo, fidelidade visual em variantes representativas e compatibilidade de navegador. A documentação de testes conserva a limitação histórica de race/checkptr nativo.
