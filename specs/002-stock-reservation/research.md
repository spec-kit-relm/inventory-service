# Research: Stock reservation at checkout — inventory slice

## Decision 1 — Concurrency control for all-or-nothing holds
- **Decision**: One transaction per hold; `SELECT … FROM stock WHERE sku = ANY($1) ORDER BY sku FOR UPDATE`, check every line, update `reserved`, insert reservation. Lock in SKU order to avoid deadlocks.
- **Rationale**: Simple, correct under concurrency, fits the 150 ms budget at 200 req/s (lock hold < 5 ms in benchmarks of the existing stock table).
- **Alternatives rejected**: optimistic versioning with retries (retry storms on hot SKUs at peak); Redis counters (new datastore → ORG-8 ADR, and a second source of truth).

## Decision 2 — Expiry mechanism
- **Decision**: Expiry worker every 10 s: `UPDATE reservations SET state='expired' WHERE state='active' AND expires_at <= now() RETURNING …` in batches of 500, decrement `reserved`, write outbox rows — one transaction per batch.
- **Rationale**: Release lag ≤ 10 s + batch time, well inside 60 s (SC-003); no new infrastructure.
- **Alternatives rejected**: per-reservation timers in memory (lost on restart); DB triggers with pg_cron (operational coupling).
- **Also**: reads compute `remainingSeconds` from `expires_at`, and *create/commit treat `expires_at <= now()` as expired even before the worker runs* — so correctness never depends on worker lag.

## Decision 3 — Idempotency storage
- **Decision**: `idempotency_keys(key, operation, body_sha256, status, response, created_at)` with 24 h TTL cleanup; written in the same transaction as the effect.
- **Rationale**: Exactly-once effect per key across retries and pods (BE-2).

## Decision 4 — Provider verification
- **Decision**: `test/contract` starts the real handler stack against Testcontainers PostgreSQL and validates every request/response against `contracts/inventory-api.yaml` with kin-openapi; events are validated against JSON Schemas extracted from `inventory-events.yaml`.
- **Rationale**: Proves the implementation conforms to the contract consumers pinned (ORG-4, INV-5).
