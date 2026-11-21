# Quickstart: verify RELM-1 in inventory-service

```bash
docker compose up -d postgres                     # local PostgreSQL 16
RESERVATION_TTL=2s go run ./cmd/inventory          # short TTL for local checks (INV-4)
```

| Scenario | Request | Expect |
|----------|---------|--------|
| RELM-1/US1/AS1 | `POST /v1/reservations` cart C, 2 × SKU-RED-MUG (5 on hand) with `Idempotency-Key` | 201, `state=active`, `remainingSeconds≈2`; availability → 3 |
| RELM-1/US1/AS2 | `POST /v1/reservations` 2 × SKU-BLUE-MUG (1 on hand) | 409 `insufficient-stock`, shortages `[{sku: SKU-BLUE-MUG, requested: 2, available: 1}]` |
| RELM-1/US1/AS3 | `GET /v1/reservations/{id}` twice | same id, decreasing `remainingSeconds`, availability unchanged |
| RELM-1/US2/AS1 | `POST /v1/reservations/{id}/commit` `{orderId}` | 200 `state=committed`; on_hand 5 → 3 |
| RELM-1/US2/AS3 | wait 3 s, commit a new reservation | 409 `reservation-expired` |
| RELM-1/US3/AS1 | wait 3 s + worker tick | availability restored; outbox row `inventory.reservation.expired` |

Automated equivalents: `go test ./...` (unit), `make integration-test`, `make contract-test`.
