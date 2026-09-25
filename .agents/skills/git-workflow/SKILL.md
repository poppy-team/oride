# Git Workflow, Branching & Evidence-Backed Review

## Purpose
Keep changes isolated, reviewable, and traceable with focused branches or worktrees, small conventional commits, secret-free history, and pull requests that carry test evidence and explicit acceptance mapping.

## Use when
- Starting scoped work: creating a feature, fix, or refactor branch or worktree from a pinned base.
- Committing: writing small conventional commits that each build and pass tests independently.
- Opening or reviewing pull requests: description, evidence, review checklist, merge strategy.
- Repairing history hygiene: rebasing a branch, splitting an oversized change, purging an accidentally committed secret.

## Do not use when
- Designing CI pipelines or merge-gate infrastructure (use `ci-cd`; this skill owns the human workflow side).
- Authoring Prumo Goals, phases, or acceptance criteria (use `goal-management`).
- Cutting versioned releases with tags and signatures (use `release-engineering`).
- Rewriting published shared history (`main`, release tags) without explicit team agreement — that is never routine.

## Required context
- Active Goal or task definition with acceptance criteria the branch must satisfy.
- Base branch and remote state: `git status`, `git log --oneline -5`, `git fetch origin` freshness.
- Repository conventions: commit message format (Conventional Commits), maximum change size (400 lines per PR), required reviewers (CODEOWNERS).
- Worktree or branch naming scheme: `feat/billing-tiered-pricing-481`, `fix/invoice-rounding-502`.

## Procedure
1. **Start isolated from a fresh base**:
   - Sync first: `git fetch origin && git checkout -b feat/billing-tiered-pricing-481 origin/main`.
   - For parallel efforts, use a linked worktree instead of stashing: `git worktree add ../prumo-fix-502 -b fix/invoice-rounding-502 origin/main`.
   - Confirm isolation with `git status --short` (clean) before writing any code.
2. **Commit small, conventional, and green**:
   - One logical change per commit, each under 150 lines, each building and passing its package tests (`go test ./internal/billing/`).
   - Format `type(scope): imperative summary` with a body explaining why, e.g. `fix(billing): round tiered tax half-up at line-item level`.
   - Allowed types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`; `refactor` commits must show zero behavior change via identical test outcomes.
   - Never commit secrets, lockfile churn from unrelated tooling, or drive-by reformatting; run `git diff --check` and `gitleaks protect --staged` before each commit.
3. **Keep the branch review-sized and current**:
   - Rebase onto `origin/main` at least daily during active work (`git pull --rebase origin main`); resolve conflicts once, promptly.
   - If the diff exceeds 400 lines, split by stacking: `feat/...-part1` merged first, part 2 branched from it.
   - Push with `--force-with-lease` only on your own feature branch after rebase; force-pushing shared branches is prohibited.
4. **Open an evidence-backed pull request**:
   - Title mirrors the leading commit; body maps each acceptance criterion to its verifying test or log line.
   - Attach evidence: CI run link, coverage delta (+1.8 percent on `internal/billing`), and manual verification notes where automation cannot reach.
   - Request the CODEOWNERS reviewers; a PR touching `internal/billing/` needs Marina Duarte's approval before merge.
5. **Merge cleanly and clean up**:
   - Merge strategy per repo policy: squash for feature branches (one commit per PR on `main`), merge-commit for release branches to preserve the trail.
   - Delete the branch and prune the worktree after merge (`git worktree remove ../prumo-fix-502 --force; git branch -d fix/invoice-rounding-502`).
   - Run the verification script `scripts/verify.sh` from the repo root; it must exit 0.

## Decision rules
- **One Branch, One Purpose**: A branch touching two unrelated concerns is split; mixed PRs are closed, not reviewed.
- **Every Commit Builds**: Each commit on a feature branch must compile and pass its package tests; `git bisect` on the branch must never land on red.
- **Conventional Format Is Mandatory**: Non-conventional messages fail review because changelog generation and SemVer inference depend on them.
- **Force-Push Is Branch-Private**: `--force-with-lease` is allowed only on unshared feature branches; shared history is append-only.
- **Secrets in History Trigger Incident Response**: A committed secret is rotated immediately and purged with `git filter-repo`, never just deleted in a follow-up commit.

## Evidence required
- Pull request record adhering to `templates/git-workflow-spec.md` with acceptance mapping and test links.
- Commit list showing conventional format, small sizes, and per-commit green status.
- `git diff --check` and staged secret-scan logs clean at each commit.
- Review approvals from CODEOWNERS reviewers recorded on the PR.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Merged change on the base branch as one traceable unit (squashed commit or merge commit per policy).
- PR documentation: scope, acceptance mapping, coverage delta, risk notes, rollback pointer.
- Clean local state: branch deleted, worktree pruned, no stash leftovers.
- History invariants intact: linear feature history or preserved release trail as policy dictates.

## Stop conditions
- PR merged with required approvals, CI green, and acceptance criteria mapped to evidence.
- Change formally rejected in review with recorded rationale and follow-up tasks filed.
- Worktree and branch cleaned up with nothing unpushed or stashed.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to the owning team if a required reviewer is unavailable for more than 2 business days; do not merge around CODEOWNERS.
- Escalate to Security immediately if a secret reaches any branch; rotate first, purge second, investigate third.
- Escalate to Lead if a merge conflict reveals two Goals editing the same contract; stop and reconcile scope before rebasing through it.
