# AGENTS.md — inventory-service

Go service that owns stock availability and checkout reservations ("holds") for Relm Commerce.
Owned by `@spec-kit-relm/backend`. **Provider** of `contracts/inventory-api.yaml` and
`contracts/inventory-events.yaml`; consumes nothing from other Relm services.
Rules that MUST hold are in `.specify/memory/constitution.md` (ORG-*, BE-*, INV-1 … INV-5).

## Commands
| Purpose | Command |
|---------|---------|
| Install deps | `go mod download` |
| Build | `go build ./...` |
| Unit tests | `go test ./...` |
| Race + coverage | `go test -race -coverprofile=cover.out ./...` |
| Integration tests (PostgreSQL via Testcontainers) | `make integration-test` *(added by RELM-1 T002)* |
| Provider verification | `make contract-test` *(added by RELM-1 T009)* |
| Lint | `go vet ./... && golangci-lint run` |
| Contract lint | `npx @redocly/cli@1 lint --config ../speckit-platform/contracts/redocly.yaml contracts/inventory-api.yaml` |
| Run locally | `RESERVATION_TTL=2s go run ./cmd/inventory` (listens on `:8081`) |

Run `go vet ./... && go test ./...` before marking any task `[X]`.

## Layout
```text
cmd/inventory/         main + wiring + config
internal/domain/       business rules — imports nothing from adapters or net/http (INV-2)
internal/httpapi/      HTTP adapter for inventory-api (stdlib ServeMux, problem+json)
internal/store/        PostgreSQL adapter + migrations           (RELM-1)
internal/events/       outbox, relay, expiry worker              (RELM-1)
contracts/             AUTHORITATIVE OpenAPI + AsyncAPI files
test/contract/         provider verification against contracts/  (RELM-1)
test/integration/      Testcontainers tests                      (RELM-1)
specs/NNN-*/           Spec Kit features (spec, plan, tasks, …)
```

## Spec Kit in this repository
- Kind: `backend` — presets `relm-org` + `relm-backend`; extensions `git`, `github`, `relm`.
- Feature numbers are product feature issue numbers: `/speckit.specify RELM-<n> …` →
  `specs/<NNN>-<short-name>/`, branch `<NNN>-<short-name>`.
- Start a cross-repo slice with `/speckit.relm.handoff RELM-<n>` (reads `../product-specs`).
- As the provider: draft contract changes during `/speckit.plan`, then open a separate contract PR.
- Implement one phase per `/speckit.implement` run; finish with `/speckit.converge` and `/speckit.relm.pr`.
- Current feature: **RELM-1** `specs/001-stock-reservation/` — tasks ready, start with Phase 1.

## Conventions
- Errors: return `domain` errors; map them to problem types only in `internal/httpapi/problems.go`.
  Problem `type` = `https://errors.relm.example/inventory/<code>`.
- Time: inject `now func() time.Time`; never call `time.Now()` inside `internal/domain`.
- Tests: table-driven; test names that verify a scenario start with its ID,
  e.g. `t.Run("RELM-1/US1/AS2 short cart holds nothing", …)`.
- Logging: `slog` with `request_id`; never log request bodies.
- SQL: lock rows in SKU order; every multi-row change in one transaction with its outbox row.

## Do not
- Change `contracts/*.yaml` in an implementation PR — contract changes are separate PRs.
- Add frameworks, ORMs or new datastores (INV-1, ORG-8).
- Use `time.Sleep` in tests; use the injected clock or `RESERVATION_TTL=2s` in integration tests.
- Edit Parts A/B of the constitution.

## Gotchas
- Availability = `on_hand − reserved`. Old code paths that read `on_hand` directly are bugs after RELM-1.
- `remainingSeconds` is computed at response time; tests must allow ±1 s.
- Create/commit treat `expires_at <= now()` as expired even if the worker has not run yet.
