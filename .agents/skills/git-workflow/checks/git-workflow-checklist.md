# Git Workflow — Verification Checklist

## 1. Branch Isolation & Naming
- [ ] Branch or worktree created from freshly fetched `origin/main`; `git status --short` was clean at start.
- [ ] Name follows scheme `type/domain-topic-id` (e.g. `feat/billing-tiered-pricing-481`); one purpose per branch.
- [ ] Parallel efforts use linked worktrees (`git worktree add`), not stash juggling or mixed working trees.
- [ ] No unrelated files in the diff; drive-by reformatting and foreign lockfile churn are absent.

## 2. Commit Discipline
- [ ] Every commit uses Conventional Commits format (`fix(billing): round tiered tax half-up at line-item level`).
- [ ] Each commit is under 150 lines, compiles, and passes its package tests independently (`git bisect` never lands on red).
- [ ] `refactor` commits show identical test outcomes before/after (zero behavior change proven, not claimed).
- [ ] `git diff --check` clean and `gitleaks protect --staged` green at every commit; no secrets staged.

## 3. Review Readiness
- [ ] Diff is under 400 lines total, or split into a stacked series with part 1 merged first.
- [ ] Branch rebased onto current `origin/main`; conflicts resolved once, promptly, with tests re-run after.
- [ ] Force-push used only with `--force-with-lease` on the author's own unshared branch.
- [ ] Shared branches (`main`, `release/*`) untouched by history rewrites.

## 4. Evidence-Backed Pull Request
- [ ] PR body maps each acceptance criterion to its verifying test or log line (no unmapped criteria).
- [ ] CI run linked and green; coverage delta reported (+1.8 percent on `internal/billing`).
- [ ] CODEOWNERS approvals recorded (Marina Duarte for `internal/billing/`); merge-block labels resolved.
- [ ] Merge strategy matches policy: squash for features, merge-commit for release branches.

## 5. Cleanup & Sign-Off
- [ ] Branch deleted and worktree pruned after merge; no stashes or unpushed commits left behind.
- [ ] PR record follows `templates/git-workflow-spec.md` with rollback pointer to the merged commit.
- [ ] Committed-secret drill known: rotate, `git filter-repo` purge, investigate — never delete-and-pretend.
- [ ] Automated verification script `scripts/verify.sh` executes with exit code 0.
