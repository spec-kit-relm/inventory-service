<!--
Sync Impact Report
- Version change: template → 1.0.0 (initial ratification)
- Part A: RELM:ORG-BASELINE v1.0.0 (synced, unchanged)
- Part B: RELM:TEAM-BACKEND v1.0.0 (synced, unchanged)
- Part C: added INV-1, INV-2, INV-3, INV-4, INV-5
- Templates: provided by presets relm-org + relm-backend (runtime resolution, no file edits)
- Deferred TODOs: none
-->
# inventory-service Constitution

> **How this constitution is layered (Relm SDD).**
> Part A is the organization baseline and Part B is the team overlay. Both are synced from
> `spec-kit-relm/speckit-platform` and MUST NOT be edited in this repository. Part C holds this
> repository's own amendments: they may add or tighten rules, never relax Parts A or B.
> Every Spec Kit command reads this file; `/speckit.analyze` treats any MUST violation as CRITICAL.

<!-- BEGIN RELM:ORG-BASELINE v1.0.0 | source: spec-kit-relm/speckit-platform/governance/org-baseline.md | DO NOT EDIT OUTSIDE speckit-platform -->
## Part A — Relm Organization Baseline (MUST)

These principles apply to every Relm repository. They are deliberately few and every one of
them names the check that enforces it. Lower layers (team overlay, repository amendments) may
**add or tighten** rules, never relax or contradict them.

### ORG-1 Spec before code
Any change that alters user-visible behavior, a public contract, or persisted data MUST trace to
a Spec Kit feature (`Feature: RELM-<n>`) or to a declared lightweight lane (`lane:bugfix`,
`lane:chore`) on the pull request.
*Check: `sdd-gates / pr-check`.*

### ORG-2 Specs describe WHAT and WHY, never HOW
`spec.md` contains no frameworks, languages, or storage choices. Every user story has
Given/When/Then acceptance scenarios and measurable success criteria. No
`[NEEDS CLARIFICATION]` marker may remain once a spec's status is `Approved` or later.
*Check: `sdd-gates / spec-lint`, `/speckit.analyze`.*

### ORG-3 Contract-first integration
Every cross-repository interaction (HTTP, events, shared schemas) is defined in a versioned
contract (OpenAPI 3.1 or AsyncAPI 3) with exactly one owning (provider) repository. Consumers pin a
version and treat vendored copies as read-only. Breaking changes require a MAJOR version and the
expand → migrate → contract sequence.
*Check: `contract-check` workflow, consumer review on contract PRs.*

### ORG-4 Tests prove the spec
Each acceptance scenario maps to at least one automated test that carries its scenario ID
(e.g. `RELM-1/US1/AS2`). Every interface touched by a feature has provider verification and
consumer contract tests, including negative cases. Tests land in the same pull request as the
behavior they verify.
*Check: CI test jobs, `/speckit.converge`, QA review.*

### ORG-5 Secure and private by default
No secrets in code, specs, prompts, or logs. Every endpoint enforces authentication and
authorization. Personal data is classified in `data-model.md` and never logged or sent to analytics.
*Check: secret scanning, CodeQL, dependency review, PR review.*

### ORG-6 Observable by design
Services emit structured logs carrying a trace ID and RED metrics (rate, errors, duration) per
endpoint or consumer. Every feature states how each success criterion is measured in production.
*Check: plan review (`Observability` section), dashboards linked before release.*

### ORG-7 Safe, reversible delivery
New behavior ships behind a feature flag, is deployed in the dependency order defined in the
product feature's `impact.md`, and has a rollback path written in `plan.md`.
*Check: plan review, release-readiness gate.*

### ORG-8 Simplicity, justified complexity
No new service, datastore, message broker, or major dependency without an accepted ADR. Any
deviation from these principles is recorded in the plan's Complexity Tracking table.
*Check: `/speckit.analyze`, architecture review.*

### ORG-9 Humans own AI output
A named human owns every pull request and reviews AI-generated code to the same standard as
hand-written code. `/speckit.analyze` reports **0 CRITICAL** findings before `/speckit.implement`,
implementation runs are scoped to one phase at a time, and `/speckit.converge` reports converged
before merge.
*Check: PR template, CODEOWNERS, `sdd-gates / pr-check`.*

### Organization governance
- **Precedence**: Org baseline (Part A) > team overlay (Part B) > repository amendments (Part C).
  If an agent finds a conflict between layers it MUST stop and report it, not choose.
- **Amendments**: only by pull request to `spec-kit-relm/speckit-platform` (`governance/`), reviewed
  by the SDD Guild plus one approver per discipline. Versioned with semantic versioning
  (MAJOR = principle removed/redefined, MINOR = principle added/expanded, PATCH = wording).
- **Propagation**: the propagation bot opens a sync PR in every repository. Repositories MUST be on
  the same MAJOR and at most one MINOR behind; `sdd-gates / constitution-check` fails otherwise.
- **Exceptions**: listed in Part C → *Exceptions* with an ADR link, an owner, and an expiry date.
<!-- END RELM:ORG-BASELINE -->



<!-- BEGIN RELM:TEAM-BACKEND v1.0.0 | source: spec-kit-relm/speckit-platform/governance/overlays/backend.md | owner: @spec-kit-relm/backend | DO NOT EDIT OUTSIDE speckit-platform -->
## Part B — Backend Team Overlay (MUST)

Applies to every backend service owned by `@spec-kit-relm/backend`. Stack-specific rules
(language, framework, layout) belong in Part C of each repository.

### BE-1 Consistent HTTP APIs
Resources are plural nouns, errors use RFC 9457 `application/problem+json`, list endpoints use
cursor pagination, and every response carries `X-Request-Id`. The provider's OpenAPI file is the
source of truth; handlers are validated against it in tests.
*Check: contract lint (Redocly ruleset `relm-api`), provider verification tests.*

### BE-2 Idempotent writes
Every non-safe operation (`POST`, `PATCH`, `DELETE`, event handlers) is idempotent: HTTP writes
accept an `Idempotency-Key` header, event consumers de-duplicate by event ID. Retries are only
allowed on idempotent operations, with bounded exponential backoff and jitter.
*Check: contract tests replay every write twice; plan review.*

### BE-3 Backward-compatible data changes
Schema migrations are forward-only and backward-compatible with the previous release
(expand → migrate → contract). Destructive steps ship in a later release than the code that
stops using the column/table.
*Check: migration lint job, plan `Rollback` section.*

### BE-4 Reliable events
Events are published through a transactional outbox, carry `eventId`, `occurredAt`,
`schemaVersion`, and `traceId`, and are documented in the provider's AsyncAPI contract.
Consumers tolerate duplicates and out-of-order delivery.
*Check: AsyncAPI lint, consumer tests with duplicate/out-of-order fixtures.*

### BE-5 Resilience budgets
Every outbound call has an explicit timeout; plans state p99 latency and error-rate budgets per
endpoint; dependencies degrade gracefully (fallback or fast failure, never unbounded waits).
*Check: plan review (`Performance & Resilience` section), load test for P1 stories.*

### BE-6 Test pyramid
Domain logic has unit tests (≥ 80 % line coverage on the domain layer), adapters have
integration tests against real dependencies in containers, and every provided interface has
provider verification. Mocks of our own services are generated from pinned contracts only.
*Check: CI coverage gate, `/speckit.converge`.*
<!-- END RELM:TEAM-BACKEND -->

## Part C — Repository Amendments (MUST)

### INV-1 Go standard library service
Go 1.23+, HTTP via `net/http` `ServeMux` patterns, logging via `log/slog` JSON with `request_id`
and `trace_id`. No web frameworks or ORMs; PostgreSQL via `pgx`. New modules need an ADR (tightens ORG-8).
*Check: `go.mod` review, `golangci-lint` `depguard` rule.*

### INV-2 Hexagonal package layout
`internal/domain` holds business rules and imports nothing from `internal/httpapi`,
`internal/store`, `internal/events`, `net/http` or `database/sql`. Adapters depend on the domain.
*Check: `golangci-lint` `depguard` (domain allow-list), PR review.*

### INV-3 Stock can never go negative
For every SKU, `available = on_hand − active_reserved ≥ 0` at all times. Enforced twice: in the
domain (all-or-nothing reservation under a row lock) and by a database `CHECK` constraint.
*Check: property test `TestStockInvariant` (random concurrent reserve/commit/expire), migration review.*

### INV-4 Reservations expire on a configurable clock
Reservation TTL comes from `RESERVATION_TTL` (default `900s`; test environments may use `2s`).
Expired units are released within 60 s by the expiry worker, which publishes
`inventory.reservation.expired` through the outbox in the same transaction (tightens BE-4).
*Check: integration test with `RESERVATION_TTL=2s`; `reservation_release_lag_seconds` alert.*

### INV-5 Provider of inventory-api and inventory-events
This repository owns `contracts/inventory-api.yaml` and `contracts/inventory-events.yaml`. Every
HTTP handler is verified against the OpenAPI document in tests (request + response validation),
and every published event against the AsyncAPI schema (provider verification, ORG-4).
*Check: `make contract-test`, `contracts / contract-check`.*

### Exceptions
None. <!-- Format: | Rule | Exception | ADR | Owner | Expires | -->

## Repository governance
- Part C is amended by pull request in this repository, approved by the CODEOWNERS of
  `.specify/memory/`, on a dedicated `governance/*` branch (never inside a feature branch).
- Parts A and B change only through the automated governance sync pull request.

**Version**: 1.0.0 | **Ratified**: 2026-11-21 | **Last Amended**: 2026-11-21
**Governance blocks**: RELM:ORG-BASELINE v1.0.0 · RELM:TEAM-BACKEND v1.0.0
