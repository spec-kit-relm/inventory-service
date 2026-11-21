---
description: "Task list for RELM-2 stock reservation — inventory-service slice"
---

# Tasks: Stock reservation at checkout — inventory slice

**Input**: Design documents from `/specs/002-stock-reservation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts (inventory-api 1.1.1, inventory-events 1.1.1 — locked)

**Tests**: Mandatory at Relm (ORG-4). Test tasks precede implementation tasks in every story phase.

**Organization**: Tasks are grouped by user story. Each phase is sized for one `/speckit.implement` run.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: US1, US2, US3 (local spec stories; central scenario IDs in test task names)

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: dependencies, config, local environment

- [ ] T001 Add `github.com/jackc/pgx/v5` and test-only `github.com/getkin/kin-openapi` to `go.mod`; add `depguard` rules for `internal/domain` to `.golangci.yml` (INV-1, INV-2)
- [ ] T002 [P] Add `docker-compose.yml` with PostgreSQL 16 and `Makefile` targets `test`, `integration-test`, `contract-test`, `lint`
- [ ] T003 [P] Add config loading in `cmd/inventory/config.go`: `DATABASE_URL`, `RESERVATION_TTL` (default 900s), `EXPIRY_INTERVAL` (default 10s)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: storage, idempotency, problems, outbox — required by every story

- [ ] T004 Migration `internal/store/migrations/0004_stock_reserved.sql`: `stock.reserved` + CHECK constraints (INV-3, BE-3)
- [ ] T005 Migration `internal/store/migrations/0005_reservations.sql`: `reservations`, `reservation_lines`, `idempotency_keys`, `outbox` per data-model.md
- [ ] T006 [P] Problem helpers in `internal/httpapi/problems.go` for `insufficient-stock`, `reservation-expired`, `idempotency-key-reused`, `validation-failed`, `temporarily-unavailable` (BE-1)
- [ ] T007 [P] Idempotency middleware in `internal/httpapi/idempotency.go` backed by `idempotency_keys` (BE-2) with unit tests for replay, different-body reuse (422), and expiry of keys
- [ ] T008 [P] Outbox writer `internal/events/outbox.go` (same-transaction insert) and relay `internal/events/relay.go` with envelope fields per AsyncAPI (BE-4)
- [ ] T009 Provider-verification harness `test/contract/harness_test.go`: start handler stack on Testcontainers PostgreSQL, validate requests/responses with kin-openapi against `contracts/inventory-api.yaml` (INV-5)
- [ ] T010 Service-token auth middleware with scopes `inventory:reserve`, `inventory:commit` in `internal/httpapi/auth.go` (ORG-5)

**Checkpoint**: storage, idempotency and provider verification ready — story work can start

---

## Phase 3: User Story 1 - Hold all cart lines when checkout starts (Priority: P1) 🎯 MVP

**Goal**: all-or-nothing holds, one active hold per cart, readable state and remaining time

**Independent Test**: create a hold for in-stock and short carts; availability reflects holds

### Tests for User Story 1 ⚠️

- [ ] T011 [P] [US1] Domain test `RELM-2/US1/AS1 hold reduces availability` in `internal/domain/reservation_test.go`
- [ ] T012 [P] [US1] Domain test `RELM-2/US1/AS2 short cart holds nothing and reports shortages`
- [ ] T013 [P] [US1] Domain test `RELM-2/US1/AS4 last unit held by another cart is not available`
- [ ] T014 [P] [US1] Domain test `INV/RELM-2/US1/AS5 new hold for same cart replaces and releases previous`
- [ ] T015 [P] [US1] Property test `TestStockInvariant` — random concurrent create/replace/commit/expire never makes available < 0 (INV-3)
- [ ] T016 [US1] Provider verification `RELM-2/US1/AS1..AS4` for `POST /v1/reservations` incl. 409 `insufficient-stock`, 422 invalid lines, replay with same `Idempotency-Key` (`INV/RELM-2/US1/AS6`) in `test/contract/reservations_test.go`
- [ ] T017 [US1] Provider verification `RELM-2/US1/AS3` for `GET /v1/reservations/{id}` (same hold, decreasing `remainingSeconds`)

### Implementation for User Story 1

- [ ] T018 [US1] Reservation aggregate and `Reserve(cart, lines, now, ttl)` in `internal/domain/reservation.go` (all-or-nothing, replace semantics)
- [ ] T019 [US1] Repository `internal/store/reservations_repo.go`: lock stock rows in SKU order, apply domain result, unique active hold per cart (research Decision 1)
- [ ] T020 [US1] Handlers `POST /v1/reservations`, `GET /v1/reservations/{id}` in `internal/httpapi/reservations.go`; availability now excludes `reserved`
- [ ] T021 [US1] Metrics `reservation_created_total{result}` and RED per route (ORG-6)

**Checkpoint**: US1 fully functional and verified against inventory-api 1.1.1

---

## Phase 4: User Story 2 - Convert a hold when the order is placed (Priority: P1)

**Goal**: idempotent commit; expired/replaced holds rejected

**Independent Test**: commit active, repeat commit, commit expired

### Tests for User Story 2 ⚠️

- [ ] T022 [P] [US2] Domain test `RELM-2/US2/AS1 commit deducts on_hand and reserved`
- [ ] T023 [P] [US2] Domain test `INV/RELM-2/US2/AS2 repeated commit for same order returns same result; other order rejected`
- [ ] T024 [P] [US2] Domain test `RELM-2/US2/AS3 commit of expired or replaced hold is rejected` (including lazy expiry before worker ran)
- [ ] T025 [US2] Provider verification for `POST /v1/reservations/{id}/commit`: 200, replay, 409 `reservation-expired`, 404

### Implementation for User Story 2

- [ ] T026 [US2] `Commit(orderID, now)` in `internal/domain/reservation.go`
- [ ] T027 [US2] Commit handler + repository transaction, outbox `inventory.reservation.committed` (BE-4)

**Checkpoint**: orders-service can integrate against the real provider in staging

---

## Phase 5: User Story 3 - Release unused holds (Priority: P2)

**Goal**: expiry worker releases within 60 s and publishes events

**Independent Test**: `RESERVATION_TTL=2s`, no commit → availability restored, event published

### Tests for User Story 3 ⚠️

- [ ] T028 [P] [US3] Integration test `RELM-2/US3/AS1 expired hold releases units within 60 s and publishes inventory.reservation.expired` (TTL 2 s) in `test/integration/expiry_test.go`
- [ ] T029 [P] [US3] Provider verification `RELM-2/US3/AS2 GET expired hold shows expired, remainingSeconds 0`
- [ ] T030 [P] [US3] Event schema test: published payloads validate against `contracts/inventory-events.yaml` (INV-5)
- [ ] T031 [US3] Race test: commit vs. expiry for the same hold — exactly one wins (spec Edge Cases)

### Implementation for User Story 3

- [ ] T032 [US3] Expiry worker `internal/events/expiry_worker.go` (batches of 500, every `EXPIRY_INTERVAL`)
- [ ] T033 [US3] Metric `reservation_release_lag_seconds` histogram

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T034 [P] Dashboard "Inventory — reservations" + alerts (invariant violations > 0, release lag p99 > 60 s) as code in `ops/dashboards/` (ORG-6)
- [ ] T035 [P] Load test profile (200 req/s, p99 ≤ 150 ms) in `test/load/reservations.js` (BE-5)
- [ ] T036 Update `AGENTS.md` Layout section and `README.md` endpoints table

---

## Phase R: Relm Verification & Handoff *(mandatory, always the last phase — ORG-3, ORG-4, ORG-9, PRD-6)*

- [ ] TR01 Provider verification and consumer contract tests pass against the locked contract versions, including negative cases
- [ ] TR02 Every acceptance scenario `RELM-2/US<k>/AS<m>` in scope has an automated test whose title starts with that ID
- [ ] TR03 Feature flag defaults to off; rollback steps from plan.md verified in a non-production environment
- [ ] TR04 Run `/speckit.converge`; resolve every appended convergence task until it reports **Converged**
- [ ] TR05 If shipped behavior or contract versions differ from the central spec, open a PR in `product-specs` updating `spec.md` / `contracts.lock.md`
- [ ] TR06 Open the pull request with the Relm PR template (Feature ID, contracts, analyze + converge evidence, scenario coverage)

---

## Dependencies & Execution Order

- Phase 1 → Phase 2 → US1 (Phase 3) → US2 (Phase 4) → US3 (Phase 5) → Phase 6 → Phase R
- US2 needs US1's aggregate; US3 needs US1 + the outbox from Phase 2.
- Within each story: tests (fail first) → domain → store → handler.
- [P] tasks in the same phase touch different files and may be delegated to sub-agents.

## Implementation Strategy

- **One `/speckit.implement` run per phase** (Relm wrapper). Suggested PRs: Phase 1–2, Phase 3, Phase 4, Phase 5–6 + R.
- MVP = Phase 3 deployed to staging: storefront and e2e-tests can start integrating.
