# inventory-service contracts (authoritative — this repo is the provider)

| Contract | Kind | Version | Tag | Consumers |
|----------|------|---------|-----|-----------|
| `inventory-api.yaml` | OpenAPI 3.1 | 1.1.1 | `inventory-api-v1.1.1` | orders-service, web-storefront, e2e-tests |
| `inventory-events.yaml` | AsyncAPI 3 | 1.1.1 | `inventory-events-v1.1.1` | — (analytics) |

## Releasing a contract version
1. Change the file in a **contract PR** (label `contract-change`) — contract only, no implementation.
2. Bump `info.version` (additive = MINOR, breaking = MAJOR via expand → migrate → contract).
3. Consumers from `product-specs/contracts/REGISTRY.md` review; `contracts / contract-check` passes.
4. After merge: `git tag inventory-api-v<version> && git push origin inventory-api-v<version>`.
5. Tell the tech lead to update `REGISTRY.md` / `contracts.lock.md` in product-specs.

## Changelog
- **inventory-api 1.1.1**, **inventory-events 1.1.1** (docs only): feature references renumbered RELM-1 → RELM-2. No interface change.
- **inventory-api 1.1.0** (RELM-2): reservations — `POST /v1/reservations`, `GET /v1/reservations/{id}`, `POST /v1/reservations/{id}/commit`.
- **inventory-events 1.1.0** (RELM-2): `inventory.reservation.expired`, `inventory.reservation.committed`.
- **inventory-api 1.0.0**: `GET /v1/items/{sku}/availability`.
