# Git Workflow, Branching Models & History Hygiene Reference Guide

## 1. Core Concepts

### 1.1 Isolation Unit: Branch vs Worktree
A branch isolates history; a worktree isolates history plus working files. One task
at a time needs a branch. Two tasks interleaved (feature plus urgent fix) need two
worktrees, because stash juggling loses state and mixes concerns. Rule: the number
of worktrees equals the number of concurrently active tasks, capped at 3 to bound
context switching.

### 1.2 Commit Atomicity and Bisectability
A commit is atomic when it compiles, passes its tests, and does exactly one thing.
Atomicity makes `git bisect` a binary search over working states instead of a walk
through rubble:

$$\text{bisect steps} \approx \log_2(N_{\text{commits}}), \quad \text{valid only if every commit builds}$$

One red commit in the range doubles debugging time and voids the guarantee. Keep
commits under 150 lines so each one stays reviewable in under 10 minutes.

### 1.3 Conventional Commits as Machine Contract
`type(scope): summary` is not style; it drives changelog generation and SemVer
inference (`feat` → minor, `fix` → patch, `feat!` → major). A non-conventional
message breaks the release tooling downstream, which is why format violations fail
review rather than earning style comments.

### 1.4 Merge Strategies and Trail Shape
- **Squash on main**: one commit per PR, linear readable history, PR number as trace key.
- **Merge-commit for release branches**: preserves the true development trail for audits.
- **Never fast-forward a feature directly**: unreviewed commits on `main` bypass every gate.

## 2. Patterns and Anti-Patterns

**Do: stack large work.** `feat/pricing-part1` (data model, merged) then
`feat/pricing-part2` (API, branched from part 1). Each PR stays under 400 lines
and reviewable; the stack preserves dependency order.

**Do: rebase daily, merge once.** `git pull --rebase origin main` keeps conflicts
small and fresh. Rebasing a week of drift in one session manufactures one giant
conflict that nobody can reason about.

**Do not: commit with `-a -m "wip"`.** Unstaged-all commits sweep in debug files,
and "wip" messages are unsearchable. Stage intentionally, summarize imperatively.

**Do not: mix refactor and behavior.** A PR that renames 40 symbols and changes
rounding logic cannot be verified: reviewers cannot tell which diff lines carry
risk. Refactor first (identical tests), behavior second.

**Do not: force-push shared branches.** Rewriting `main` orphans every open PR and
invalidates every checked-out clone. `--force-with-lease` on your own feature
branch is the only acceptable force-push.

## 3. Command Configuration Example

```bash
# Start isolated work from a fresh base (branch + worktree variants)
git fetch origin
git checkout -b feat/billing-tiered-pricing-481 origin/main
git worktree add ../prumo-fix-502 -b fix/invoice-rounding-502 origin/main

# Commit with hygiene gates: whitespace errors + staged secret scan
git diff --check
gitleaks protect --staged
git commit -m "fix(billing): round tiered tax half-up at line-item level"

# Stay current and keep the change review-sized
git pull --rebase origin main
git log --oneline --stat -8   # each commit builds; total diff under 400 lines
git push -u origin feat/billing-tiered-pricing-481

# Cleanup after merge
git worktree remove ../prumo-fix-502 --force
git branch -d fix/invoice-rounding-502
git worktree prune
```
