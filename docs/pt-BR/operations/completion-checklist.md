# go-image-optimizer: Checklist de conclusão do repositório

Revisão estática em 2026-10-10; sem execução runtime. Itens marcados indicam documentação concluída, não funcionalidades entregues ou lacunas encerradas.

- [x] Inventário: documentação e código/configuração/testes inspecionados; inventário original no manifest da biblioteca.
- [x] Reestruturação: categorias equivalentes por idioma; IDs/histórico ADR preservados; arquivos exigidos por ferramentas mantidos.
- [x] Revisão de conteúdo: mecanismos, contratos, alternativas, falhas e limites de evidência explicados.
- [x] Cobertura bilíngue: páginas mantidas equivalentes en/pt-BR; históricos identificados explicitamente.
- [x] Navegação: README e catálogo por idioma alcançam cada documento mantido.
- [x] Cobertura Engineering Library: explicações completas e respostas substanciais preservadas/ampliadas.
- [x] Validação de links/anchors/numeração: verificação final sem erros; 18 warnings de links restritos de fixtures registrados no relatório da biblioteca.

## Evidência inspecionada

- [backend/internal/infrastructure/http/router.go](../../../backend/internal/infrastructure/http/router.go)
- [backend/internal/infrastructure/http/server.go](../../../backend/internal/infrastructure/http/server.go)
- [backend/internal/application/imagecompression/usecase.go](../../../backend/internal/application/imagecompression/usecase.go)
- [backend/internal/infrastructure/imaging/limits.go](../../../backend/internal/infrastructure/imaging/limits.go)
- [frontend/app/api/images/forward-image-request.ts](../../../frontend/app/api/images/forward-image-request.ts)
- [docker-compose.yml](../../../docker-compose.yml)

## Cobertura de entrevista e contraparte na biblioteca

Compressão/resize/conversão/preview; interfaces/CGO; multipart/API/proxy; pixels, formatos, alpha/orientação/animação; cleanup/cancelamento; testes, recursos e limites de rollout.

[Self-contained dossier / Dossiê](../../../../engineering-library/docs/pt-BR/architecture/go-image-optimizer.md) | [Interview / Entrevista](../../../../engineering-library/docs/pt-BR/interviews/go-image-optimizer.md)

## Aplicabilidade das categorias

| Categoria | Tratamento / justificativa |
| --- | --- |
| database | Não aplicável: sem armazenamento persistente. |
| payments | Não aplicável: pagamentos ausentes. No Payment Lab, conciliação/Stripe são apenas direção futura. |
| integrations | Next.js→Go e dependências nativas em api/architecture/adr; sem chamadas entre repositórios. |
| deployment | Dependências runtime e limites de atualização/rollback em operations/docker. |
| observability | Diagnóstico slog/health em operations. |
| benchmarks | Não aplicável: sem medições de desempenho verificadas adequadas a gráficos; nenhuma executada. |

Categorias presentes no índice contêm conteúdo mantido; categorias tratadas em outros locais não recebem pastas vazias. ADR/comparações conservam histórico; nenhuma motivação ou data histórica nova foi inventada.

## Lacunas restantes dependentes de evidência

Interrupção nativa, pico concorrente, compatibilidade visual/navegador, limitação histórica race/checkptr e implantação produtiva.
