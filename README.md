# inventory-service

**Stock availability and checkout reservations for Relm Commerce.** Go 1.23, stdlib `net/http`.

Owner: `@spec-kit-relm/backend` · Spec Kit project, kind **backend** (presets `relm-org` +
`relm-backend`, extensions `git`, `github`, `relm`) · Provider of **inventory-api** and
**inventory-events**

## Interfaces

| Contract | Version | Consumers | Notes |
|----------|---------|-----------|-------|
| [`contracts/inventory-api.yaml`](contracts/inventory-api.yaml) | 1.1.1 | orders-service, web-storefront, e2e-tests | 1.1.0 adds reservations (RELM-2) |
| [`contracts/inventory-events.yaml`](contracts/inventory-events.yaml) | 1.1.1 | — | reservation expired/committed events |

## Run it

```bash
go test ./...
go run ./cmd/inventory            # :8081
curl localhost:8081/v1/items/SKU-RED-MUG/availability
```

Today's code implements inventory-api **1.0.0** (availability). RELM-2 (below) adds the
reservation endpoints — the tasks are ready and implementation has not started yet.

## How we work in this repository — in order

1. **Read the rules**: [`.specify/memory/constitution.md`](.specify/memory/constitution.md)
   (Part A company ORG-*, Part B backend BE-*, Part C this repo INV-1 … INV-5) and
   [`AGENTS.md`](AGENTS.md) (commands, layout, conventions).
2. **Receive a feature**: `/speckit.relm.handoff RELM-<n>` reads your slice from `../product-specs`.
3. **Specify your slice**: `/speckit.specify RELM-<n> <short-name>: …` → `specs/NNN-short-name/`
   (same number and branch as every other repo), then `/speckit.clarify`.
4. **Plan, contract first**: `/speckit.plan`. If the feature changes our contracts, promote the
   draft to `contracts/`, bump the version, and open a **contract PR** for consumers (Gate 2).
5. **Tasks + analyze**: `/speckit.tasks`, then `/speckit.analyze` until 0 CRITICAL;
   `/speckit.github.taskstoissues`; plan PR (Gate 3).
6. **Implement**: `/speckit.implement Phase <k>` — one phase per run, review, test, PR.
7. **Converge + PR**: `/speckit.converge` until Converged; `/speckit.relm.pr` for the description.
8. **Release** in the order of the product feature's `impact.md` (this service is usually first).

Bugfixes and chores skip steps 2–5 — see the
[work lanes](https://github.com/spec-kit-relm/sdd-handbook/blob/main/docs/03-work-lanes.md).

## Current feature

| Feature | Directory | Role | State |
|---------|-----------|------|-------|
| RELM-2 Stock reservation at checkout | [`specs/002-stock-reservation/`](specs/002-stock-reservation/) | provider (first in rollout) | spec Approved · contracts locked · [tasks](specs/002-stock-reservation/tasks.md) ready — **next: `/speckit.implement Phase 1`** |

## Gates on every PR

`sdd / sdd-gates` (agent files, constitution drift, spec lint, PR traceability),
`contracts / contract-check` (lint, version bump, breaking changes), `ci / test` (vet, race tests,
domain coverage ≥ 80 %), one `@spec-kit-relm/backend` approval.
