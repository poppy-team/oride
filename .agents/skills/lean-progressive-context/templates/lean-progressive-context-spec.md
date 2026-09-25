# Progressive Context Plan Specification

## 1. Run Identity
- **Task**: Fix token-refresh race in auth service (Goal G-118, Task T-342)
- **Budget envelope**: total 60,000 tokens, per-stage cap 8,000 tokens, max 5 stages
- **Entry-point hints**: auth/refresh.go, bug report BR-2091, auth spec docs/security/service-tokens.md
- **Sufficiency bar**: can name the files to change and the test command proving the fix

## 2. Expansion Ledger

| Stage | Hypothesis | Files loaded | Stage cost | Running total |
|---|---|---|---|---|
| 1 | Minimal set suffices for orientation | bug report BR-2091, auth/refresh.go, service-tokens.md | 3,100 | 3,100 |
| 2 | Race is between refresh and revoke paths | auth/revoke.go, auth/refresh_test.go, auth/revoke_test.go | 2,400 | 5,500 |
| stop | Sufficiency reached at K=2 | Change refresh.go mutex scope; run go test ./auth -race | 0 | 5,500 |

## 3. Working Context Capsule (Stage 2)
- **Goal**: fix refresh/revoke race from BR-2091.
- **Decisions**: refresh path holds no lock across the DB write; revoke deletes the row mid-refresh.
- **Evidence**: refresh_test.go has no concurrent test; race visible under -race with 50 goroutines.
- **Open**: whether revoke should tombstone instead of delete (deferred to reviewer).
- **Next layer**: none needed; proceed to fix.

## 4. Budget Outcome
- **Spent**: 5,500 of 60,000 tokens across 2 of 5 stages.
- **Unspent**: 54,500 tokens returned to the run envelope.
- **Overrides**: none.

## 5. Verification Evidence
- [ ] Ledger reconciles: 3,100 + 2,400 = 5,500 reported spend.
- [ ] Capsule present, 11 lines, covers goal/decisions/evidence/open/next.
- [ ] `scripts/verify.sh` exits 0.
