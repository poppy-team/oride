# Lean Progressive Context — Verification Checklist

## 1. Budget Declaration
- [ ] Total token budget, per-stage cap, and max expansion stages are recorded before any file is loaded.
- [ ] Default envelope (60,000 total / 8,000 per stage / 5 stages) is used unless the Task justifies otherwise in writing.
- [ ] Budget overrides carry a written rationale, not a silent increase.

## 2. Minimal Working Set
- [ ] First load contains only the task statement, entry-point file, and directly referenced spec.
- [ ] No directory-wide glob or recursive read appears in the first layer.
- [ ] Token cost of the first layer is measured and logged, not estimated.

## 3. Hypothesis-Driven Expansion
- [ ] Every expansion layer states a falsifiable hypothesis about what it expects to find.
- [ ] Each loaded file is charged against the ledger with per-file and running-total costs.
- [ ] No layer loads files unrelated to its stated hypothesis.
- [ ] Expansion halts at the first sufficient layer; further loading is justified or flagged.

## 4. Capsules & Handoff
- [ ] Each completed stage produces a Working Context Capsule of at most 15 lines.
- [ ] Capsules record goal, key decisions, open questions, and the next layer if needed.
- [ ] Handoffs carry capsules, never raw unbounded file lists.

## 5. Verification Gates
- [ ] Ledger totals reconcile: sum of per-file costs equals the reported spend.
- [ ] Sufficiency note names the answering layer and the evidence that stopped expansion.
- [ ] Automated verification script `scripts/verify.sh` executes with exit code 0.
