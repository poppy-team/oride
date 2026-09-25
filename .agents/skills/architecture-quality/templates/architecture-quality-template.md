# Architecture Quality Review — Billing Service Extraction

## 1. Metadata
- **Skill**: architecture-quality (Architecture Quality)
- **Date**: 2026-09-15
- **Author / Agent**: billing-platform-team
- **Target Goal / Phase**: GOAL-042 billing-service-extraction

## 2. Executive Summary
Reviewed the proposed extraction of billing logic from the `monolith-core` package into a new `billing` bounded context. Found one dependency cycle (`billing → notifications → billing`) and one boundary leak (ORM entities crossing the HTTP boundary). Both resolved by introducing a `BillingEvents` port interface and `InvoiceDTO` mappers. Fitness functions added to CI to pin the new boundaries.

## 3. Inputs & Scope
- **Inputs Evaluated**: ADR-011 (billing extraction), module dependency graph (`go list -deps`), `Invoice`/`Charge` contract definitions
- **Artifacts Modified**: `internal/billing/*` (new), `internal/notifications/sender.go`, `api/openapi/billing.yaml`

## 4. Key Findings & Implementation Details
- **Cycle broken**: `notifications` no longer imports `billing`; both depend on the `billingevents.Publisher` port. Verified acyclic with `deptrac` (0 cycles, was 1).
- **Boundary sealed**: HTTP handlers map `Invoice` domain structs to `InvoiceDTO` at the edge; DB models stay inside the repository layer. Added ArchUnit-style test `TestBillingBoundaryIsolation` asserting `internal/billing/domain` imports nothing from `api/` or `storage/`.
- **God module split**: `billing.go` (1,900 lines) decomposed into `invoice.go`, `charge.go`, `refund.go` behind the `BillingService` facade; largest file now 420 lines.
- **Stable interface**: `BillingService` versioned as `v1` with a deprecation policy (2 minor versions notice) to protect the checkout caller.

## 5. Verification & Evidence
- **Evidence Type**: review, test
- **Test Results**: Passed — `go test ./internal/billing/...` 47/47 green; `deptrac` 0 violations; boundary test green
- **Static Analysis Status**: Pass — `go vet` clean; no new import cycles

## 6. Next Steps & Handoff
- Wire `BillingService v1` into checkout flow (downstream task GOAL-043).
- Add latency SLI on `Charge` path before cutover; owner: billing-platform-team.
