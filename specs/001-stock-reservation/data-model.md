# Data Model: Stock reservation at checkout — inventory slice

## Tables

| Table | Columns | Constraints |
|-------|---------|-------------|
| `stock` (existing) | `sku text pk`, `on_hand int`, **`reserved int not null default 0`** | **`CHECK (on_hand - reserved >= 0)`**, `CHECK (reserved >= 0)` (INV-3) |
| `reservations` | `id uuid pk`, `cart_id uuid`, `state text`, `created_at timestamptz`, `expires_at timestamptz`, `order_id uuid null`, `updated_at` | unique partial index `(cart_id) WHERE state='active'` (one active per cart) |
| `reservation_lines` | `reservation_id uuid fk`, `sku text fk`, `quantity int` | `CHECK (quantity BETWEEN 1 AND 99)`, pk `(reservation_id, sku)` |
| `idempotency_keys` | `key uuid`, `operation text`, `body_sha256 bytea`, `status int`, `response jsonb`, `created_at` | pk `(key, operation)`; purge after 24 h |
| `outbox` | `id uuid pk`, `event_type text`, `payload jsonb`, `occurred_at`, `published_at null` | index on `published_at IS NULL` |

## Reservation state machine

```text
            create (all lines available)
   ┌──────────────────────────────────────┐
   │                                      ▼
 (none) ──────────────────────────────► active ──commit(order O)──► committed   (terminal; commit(O) again = same result)
                                          │  │
                    new hold for same cart│  │expires_at reached (worker or lazy check)
                                          ▼  ▼
                                    replaced  expired                (terminal; commit → 409 reservation-expired)
```

| Transition | Stock effect | Event |
|------------|--------------|-------|
| create | `reserved += qty` per line | — |
| replace | old: `reserved -= qty`; new: `reserved += qty` (one transaction) | — |
| commit | `on_hand -= qty`, `reserved -= qty` | `inventory.reservation.committed` |
| expire | `reserved -= qty` | `inventory.reservation.expired` |

## Personal data
None. `cart_id`, `order_id` are opaque identifiers (ORG-5).
