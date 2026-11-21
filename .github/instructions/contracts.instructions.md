---
applyTo: "contracts/**"
---
These files are AUTHORITATIVE contracts owned by inventory-service (INV-5, ORG-3).
- Any change bumps `info.version` (additive = MINOR, breaking = MAJOR with expand → migrate → contract).
- Describe semantics in `description` fields: idempotency, retry safety, problem types, ordering.
- A contract change is its own PR (label `contract-change`) reviewed by every consumer in
  product-specs/contracts/REGISTRY.md — never mixed with implementation.
