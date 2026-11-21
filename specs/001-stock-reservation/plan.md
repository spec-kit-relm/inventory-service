# Implementation Plan: Stock reservation at checkout — inventory slice

**Branch**: `001-stock-reservation` | **Date**: 2026-11-19 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-stock-reservation/spec.md`

## Summary

Add reservations ("holds") to inventory-service: an all-or-nothing hold of all cart lines under a
row-level lock, one active hold per cart (replace releases atomically), idempotent create/commit,
and an expiry worker that releases unused holds within 60 s and publishes
`inventory.reservation.expired` through a transactional outbox. Contract: inventory-api 1.0.0 →
**1.1.0** and inventory-events → **1.1.0** (additive, locked in product-specs).

## Technical Context

**Language/Version**: Go 1.23 (INV-1)

**Primary Dependencies**: stdlib `net/http`, `log/slog`; `jackc/pgx/v5` (PostgreSQL); `getkin/kin-openapi` (test-only, provider verification)

**Storage**: PostgreSQL 16 — new tables `reservations`, `reservation_lines`, `idempotency_keys`, `outbox`; `stock` gains `reserved` column with `CHECK (on_hand - reserved >= 0)` (INV-3)

**Testing**: `go test` (unit + property tests), Testcontainers-Go for PostgreSQL integration tests, kin-openapi request/response validation (INV-5)

**Target Platform**: Linux container on the Relm Kubernetes platform

**Project Type**: web-service

**Performance Goals**: 200 hold requests/s sustained at peak (Black-Friday profile × 2)

**Constraints**: p99 ≤ 150 ms for `POST /v1/reservations` and `/commit`; expiry release lag p99 ≤ 60 s

**Scale/Scope**: ~40 k SKUs, ≤ 50 lines per reservation, ~1 M reservations/day at peak

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Rule | How this plan complies | Status |
|------|------------------------|--------|
| ORG-1 Spec before code | Feature RELM-1, local spec approved | ✅ |
| ORG-2 WHAT/WHY in spec | Local spec is technology-free; tech only here | ✅ |
| ORG-3 Contract-first | inventory-api 1.1.0 + inventory-events 1.1.0 promoted and locked before tasks | ✅ |
| ORG-4 Tests prove the spec | Scenario-ID tests per story; provider verification for every new operation | ✅ |
| ORG-5 Secure by default | Service-token auth, scope `inventory:reserve`; no personal data (cartId is opaque) | ✅ |
| ORG-6 Observable | metrics `reservation_created_total`, `reservation_release_lag_seconds`, RED per endpoint | ✅ |
| ORG-7 Safe delivery | Endpoints are additive and unused until storefront flag is on; rollback below | ✅ |
| ORG-8 Simplicity | No new datastore or broker: outbox table + existing Kafka relay; no ADR needed | ✅ |
| ORG-9 Humans own AI output | One phase per implement run; analyze 0 CRITICAL recorded in plan PR | ✅ |
| BE-1 HTTP API style | problem+json types `insufficient-stock`, `reservation-expired`, `idempotency-key-reused` | ✅ |
| BE-2 Idempotent writes | `idempotency_keys` table, 24 h window, body hash comparison | ✅ |
| BE-3 Compatible migrations | Additive tables/column with default 0; no destructive step | ✅ |
| BE-4 Reliable events | Outbox in same transaction; envelope per AsyncAPI | ✅ |
| BE-5 Budgets | See budgets table | ✅ |
| BE-6 Test pyramid | Domain unit + property, PostgreSQL integration, provider verification | ✅ |
| INV-1 … INV-5 | stdlib only + pgx; domain isolated; CHECK constraint; `RESERVATION_TTL`; contract tests | ✅ |

## Project Structure

### Documentation (this feature)

```text
specs/001-stock-reservation/
├── plan.md              # This file
├── research.md          # Decisions: locking strategy, expiry worker, idempotency storage
├── data-model.md        # Tables, states, transitions
├── quickstart.md        # How to run and verify the scenarios locally
├── contracts/           # Draft that was promoted to /contracts (kept for history)
└── tasks.md             # /speckit.tasks output
```

### Source Code (repository root)

```text
cmd/inventory/                 main, wiring, config (RESERVATION_TTL)
internal/domain/               stock.go, reservation.go (+ _test, property tests)   — no adapters (INV-2)
internal/httpapi/              server.go, reservations.go, problems.go, idempotency.go
internal/store/                postgres.go, reservations_repo.go, migrations/*.sql
internal/events/               outbox.go, relay.go, expiry_worker.go
contracts/                     inventory-api.yaml, inventory-events.yaml (authoritative)
test/contract/                 provider verification against contracts/*.yaml
test/integration/              Testcontainers PostgreSQL tests
```

**Structure Decision**: Single Go service with hexagonal packages (INV-2).

## Complexity Tracking

> No constitution violations to justify.

---

## Relm Delivery Sections *(mandatory — ORG-3, ORG-5, ORG-6, ORG-7)*

### Contracts & Versions (ORG-3)

| Interface | Role | Contract file | Locked version | Change | Consumers / provider |
|-----------|------|---------------|----------------|--------|----------------------|
| Inventory API — availability + reservations | provides | `contracts/inventory-api.yaml` | 1.1.0 | additive | orders-service, web-storefront, e2e-tests |
| Inventory events — reservations channel | provides | `contracts/inventory-events.yaml` | 1.1.0 | additive | — (analytics) |

### Observability (ORG-6)

| Success criterion (from spec) | Signal (metric / log / trace) | Dashboard or alert |
|-------------------------------|-------------------------------|--------------------|
| SC-001 never negative | `stock_invariant_violations_total` (must stay 0); DB CHECK violations logged as `error` | alert on > 0 (T034) |
| SC-002 p95 ≤ 150 ms | `http_server_duration_seconds{route="/v1/reservations"}` | "Inventory — reservations" dashboard (T034) |
| SC-003 release ≤ 60 s | `reservation_release_lag_seconds` histogram | alert p99 > 60 s for 10 min (T034) |

### Feature flag, rollout & rollback (ORG-7)

- **Flag**: none in this service — endpoints are additive; the shopper-facing flag `checkout.stock-reservation` lives in web-storefront.
- **Rollout position** (from impact.md): 1 of 4 — first.
- **Rollback**: redeploy previous version; active holds stop being enforced and the expiry worker stops. Safe only after storefront flag is off (consumers first). Tables stay (forward-only migrations).

### Security & privacy (ORG-5)

- **AuthN / AuthZ**: service token required; `inventory:reserve` scope for create/read, `inventory:commit` for commit (orders-service only).
- **Personal data touched**: none. `cartId` and `orderId` are opaque UUIDs.
- **Secrets / config**: `DATABASE_URL` (secret), `RESERVATION_TTL` (config, default 900s).

### Backend: idempotency & retries (BE-2)

| Operation | Idempotency mechanism | Retry policy | Duplicate behavior |
|-----------|----------------------|--------------|--------------------|
| POST /v1/reservations | `Idempotency-Key` + SHA-256 of body, 24 h | clients may retry on 503 / timeout | original 201 + body; different body → 422 `idempotency-key-reused` |
| POST /v1/reservations/{id}/commit | `Idempotency-Key`; plus natural key (reservation, orderId) | orders-service retries 3× with backoff | 200 with same body |
| Expiry worker | state transition `active → expired` guarded by `WHERE state='active'` | runs every 10 s | no-op on already expired |

### Backend: data changes & migrations (BE-3)

| Change | Migration | Backward-compatible with previous release? | Contract step (later release) |
|--------|-----------|---------------------------------------------|-------------------------------|
| `stock.reserved int not null default 0` + CHECK | `0004_stock_reserved.sql` | yes — old code ignores the column | none |
| new tables `reservations`, `reservation_lines`, `idempotency_keys`, `outbox` | `0005_reservations.sql` | yes — unused by old code | none |

### Backend: events (BE-4)

| Event | Direction | AsyncAPI channel | Outbox? | Consumer de-dup key |
|-------|-----------|------------------|---------|---------------------|
| inventory.reservation.expired | publishes | inventory.reservations.v1 | yes | eventId |
| inventory.reservation.committed | publishes | inventory.reservations.v1 | yes | eventId |

### Backend: performance & resilience budgets (BE-5)

| Endpoint / consumer | p99 latency budget | Error budget | Timeouts on outbound calls | Degradation behavior |
|---------------------|--------------------|--------------|----------------------------|----------------------|
| POST /v1/reservations | 150 ms | 0.1 % | DB statement 100 ms, lock wait 50 ms | 503 `temporarily-unavailable` + Retry-After: 1 |
| POST /v1/reservations/{id}/commit | 150 ms | 0.1 % | DB statement 100 ms | 503 + Retry-After |
| GET /v1/reservations/{id} | 50 ms | 0.1 % | DB 30 ms | 503 |
| expiry worker | lag p99 60 s | — | batch of 500, 2 s per batch | backlog alert |
