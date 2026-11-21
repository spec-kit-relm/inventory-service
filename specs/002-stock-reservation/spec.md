# Feature Specification: Stock reservation at checkout — inventory slice

**Feature Branch**: `002-stock-reservation`

**Created**: 2026-11-18

**Status**: Approved

**Input**: User description: "RELM-2 stock-reservation: inventory-service slice from the handoff brief — hold all cart lines at checkout start for a fixed time (all-or-nothing), expose the hold's state and remaining time, convert a hold into a permanent deduction when an order is placed, and release unused holds within 60 seconds of expiry."

## Clarifications

### Session 2026-11-18

- Q: Who calls the hold operations? → A: The storefront (create, read) and orders-service (commit). Both are authenticated services; shoppers never call inventory-service directly.
- Q: What does "replace the previous hold" mean for stock? → A: Units of the replaced hold are released immediately, in the same operation that creates the new hold.
- Q: Can a committed hold be committed again? → A: Yes, by the same order: it returns the same result (retry safety). A different order is rejected.
- Q: Must the 15-minute duration be configurable? → A: Yes for test environments only; production stays at 15 minutes (central FR-002).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Hold all cart lines when checkout starts (Priority: P1)

The storefront asks inventory to hold a cart's lines; inventory holds all of them or none and
reports what is short.

**Why this priority**: Core of RELM-2 (central US1); every other repository depends on it.

**Independent Test**: Create a hold for an in-stock cart, verify availability for other carts drops
accordingly; create a hold for a short cart, verify nothing is held and shortages are reported.

**Acceptance Scenarios**:

1. **Given** SKU A with 5 available, **When** a hold for 2 units of A is requested for cart C, **Then** the hold is active for 15 minutes and A shows 3 available. *[RELM-2/US1/AS1]*
2. **Given** SKU A with 1 available, **When** a hold for 2 units of A is requested, **Then** nothing is held and the response lists A as short with 1 available. *[RELM-2/US1/AS2]*
3. **Given** an active hold for cart C, **When** the hold is read again, **Then** the same hold and the correct remaining time are returned and availability is unchanged. *[RELM-2/US1/AS3]*
4. **Given** the only unit of A is held by cart C, **When** cart D requests a hold for A, **Then** nothing is held for D and A is reported short. *[RELM-2/US1/AS4]*
5. **Given** an active hold for cart C, **When** a new hold with different lines is requested for C, **Then** the previous hold is replaced and its units are released in the same operation.
6. **Given** a hold request, **When** it is repeated with the same idempotency key, **Then** the original result is returned and no additional units are held.

---

### User Story 2 - Convert a hold when the order is placed (Priority: P1)

orders-service commits the hold when an order is placed; the units become a permanent deduction.

**Why this priority**: Central US2 depends on it; without commit, holds would expire and stock would be resold.

**Independent Test**: Commit an active hold and verify on-hand stock decreases permanently; repeat the commit; commit an expired hold.

**Acceptance Scenarios**:

1. **Given** an active hold, **When** it is committed for order O, **Then** the held units are permanently deducted and the hold is marked committed. *[RELM-2/US2/AS1]*
2. **Given** a hold committed for order O, **When** the commit for O is repeated, **Then** the same result is returned and nothing else changes.
3. **Given** an expired or replaced hold, **When** a commit is attempted, **Then** it is rejected as expired and nothing is deducted. *[RELM-2/US2/AS3]*

---

### User Story 3 - Release unused holds (Priority: P2)

Holds that are not committed within their lifetime release their units automatically.

**Why this priority**: Prevents abandoned checkouts from locking stock (central US3).

**Independent Test**: With a short test lifetime, create a hold, do not commit it, and verify the units return within 60 seconds and an expiry notification is published.

**Acceptance Scenarios**:

1. **Given** an active hold that is not committed, **When** its lifetime ends, **Then** its units are available again within 60 seconds and an expiry event is published. *[RELM-2/US3/AS1]*
2. **Given** an expired hold, **When** it is read, **Then** it shows the expired state and zero remaining time. *[RELM-2/US3/AS2]*

---

### Edge Cases

- Expiry and commit race: exactly one wins; the loser sees a consistent state (committed, or rejected as expired).
- Two holds requested concurrently for the last unit: exactly one succeeds.
- A warehouse adjustment lowers on-hand below the held quantity: existing holds stay valid; availability is reported as 0, never negative.
- More than 50 lines or quantity above 99 per line: rejected as invalid input.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Holds MUST be all-or-nothing across all lines of a request.
- **FR-002**: A hold MUST last 15 minutes in production; the lifetime MUST be configurable in test environments.
- **FR-003**: Availability MUST exclude units in active holds and MUST never be negative.
- **FR-004**: A cart MUST have at most one active hold; a new hold for the same cart replaces and releases the previous one atomically.
- **FR-005**: Create and commit MUST be safe to retry with the same idempotency key.
- **FR-006**: Commit MUST permanently deduct units and be idempotent per order.
- **FR-007**: Commit of an expired or replaced hold MUST be rejected without deducting units.
- **FR-008**: Expired holds MUST release units within 60 seconds and publish an expiry event.
- **FR-009**: Reading a hold MUST return its state and remaining time.

### Key Entities *(include if feature involves data)*

- **Reservation (hold)**: cart, lines (SKU, quantity), state, created/expires time, committing order.
- **Stock ledger**: per SKU on-hand units and actively reserved units.
- **Outbox entry**: an event waiting to be published, written in the same transaction as the state change.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: No SKU ever reports negative availability or is committed beyond on-hand (central SC-001).
- **SC-002**: 95 % of hold requests complete within 150 ms at the service boundary (supports central SC-002's 1 s end-to-end).
- **SC-003**: 99 % of expired holds release their units within 60 seconds (central SC-003).

## Assumptions

- Consumers authenticate with service tokens; per-shopper authorization stays in the storefront.
- No Relm service consumes the reservation events in RELM-2; they are published for analytics and future features.

---

## Relm Traceability *(mandatory — ORG-1, PRD-1, QA-2)*

**Feature ID**: RELM-2
**Central spec**: [product-specs/specs/002-stock-reservation/spec.md](https://github.com/spec-kit-relm/product-specs/blob/main/specs/002-stock-reservation/spec.md)
**Repository role**: provider
**Contracts provided**: inventory-api@1.1.1, inventory-events@1.1.1
**Contracts consumed**: none
**Product stories covered**: RELM-2 US1 (AS1–AS4), US2 (AS1, AS3), US3 (AS1, AS2)

**Status lifecycle**: Draft → In Review → Approved → Implemented → Released

### Scenario IDs
Central scenario IDs are shown in brackets after each shared scenario. Repository-only scenarios
(US1 AS5–AS6, US2 AS2) use local IDs `INV/RELM-2/US<k>/AS<m>` in test titles.
