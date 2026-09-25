# Refactoring Delivery Record

## Metadata
- Skill: Refactoring
- Date: 2026-09-23
- Author: refactoring-agent
- Goal: REF-07 — isolate invoice calculation without changing public behavior

## Baseline
- Characterization tests: 18 passing
- Public API: `InvoiceService.calculate(invoice, adjustments) -> Decimal`
- Golden output: `invoice-2026-09.fixture.json`

## Transformations
1. Extracted `normalize_adjustments` from the calculation method.
2. Extracted `validate_currency` and kept its validation order unchanged.
3. Replaced the repeated discount threshold with `ENTERPRISE_DISCOUNT_THRESHOLD`.
4. Left transport, persistence, and logging code outside the refactor boundary.

## Verification
- Each transformation kept the characterization suite green.
- The before/after fixture diff is empty.
- Complexity of the touched method fell from 14 to 7.
- No public signature or error code changed.

## Follow-Up
- A separate issue tracks replacing the legacy tax adapter; it is not part of this refactor.
