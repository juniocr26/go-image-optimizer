# Documentation catalog: go-image-optimizer

[Project introduction](../../README.md) | [Other language](../pt-BR/index.md)

Reading path: purpose/setup → architecture/contracts → security/failures → testing/evidence → operations. For the library, use catalog/dossier → project interview → foundations → failure follow-ups. Historical milestone/refactor/handoff records describe their original dates; current review/checklist explains present scope.

## architecture

Components, flows and implemented boundaries.

- [Architecture](architecture/overview.md)

## adr

Recorded decisions and consequences; identifiers/history preserved.

- [ADR 001: Native Image Codecs](adr/001-native-image-codecs.md)

## guides

Setup, prerequisites and reading procedures.

- [Go Image Optimizer](guides/project-guide.md)

## testing

Test strategy, inventories and dated evidence.

- [Test Documentation](testing/strategy.md)
- [Documentation validation — 2026-10-06](testing/verification.md)

## docker

Local images, services, mounts and configuration.

- [Docker development setup and recovery](docker/development.md)
- [Docker](docker/runtime.md)

## api

Exposed contracts and client behavior.

- [Image API and proxy contract](api/contracts.md)

## operations

Diagnosis, recovery, review checklists and historical records.

- [go-image-optimizer: Repository completion checklist](operations/completion-checklist.md)
- [Deployment, diagnostics and evidence](operations/deployment-and-diagnosis.md)

## security

Authentication, authorization, data and resource boundaries.

- [Resource and native-code boundaries](security/resource-boundaries.md)

Category applicability and omissions are justified in the [completion checklist](operations/completion-checklist.md). No empty category is created.
