# CI/CD Pipeline Design & Hardening Reference Guide

## 1. Core Concepts

### 1.1 The Reproducibility Contract
A pipeline is reproducible when `clone at SHA + run workflow = same verdict`, regardless of
runner history. Three inputs determine the verdict: the pinned environment (runner image,
action SHAs, toolchain), the locked dependencies (lockfiles), and the source tree. Any
unpinned input is a hidden variable that converts deterministic gates into lottery tickets.

### 1.2 Fail-Fast Economics
Pipeline cost per broken commit follows:

$$\text{Cost} = \sum_{\text{stages run}} (\text{runner-minutes} \times \text{rate}) + \text{developer wait time}$$

Running a 9-minute test suite on a commit that fails `gofmt` in 20 seconds wastes both.
Order gates by ascending cost: format and lint first, unit tests second, integration and
scans last. Cancel downstream stages on Stage 1 failure with `needs:` chains.

### 1.3 Cache Correctness Model
A dependency cache is a function `key(lockfile_hash, os, arch) -> objects`. Correctness
requires: the key changes whenever inputs change (hash the lockfile, not the branch name),
hits are bounded in age (7-day TTL policy with explicit purge), and a cold miss still
produces a green build (just slower). A cache that is required for green is not a cache;
it is an undeclared input.

### 1.4 Secret Blast Radius
$$\text{Blast radius} = \text{jobs with access} \times \text{log retention} \times \text{artifact lifetime}$$

Minimize all three: scope secrets to one job, keep logs at the minimum retention that
satisfies audit (30 days), and never upload secret-bearing files as artifacts. OIDC
federation reduces the standing-credential term to zero for cloud and registry access.

## 2. Patterns and Anti-Patterns

**Do: pin actions by SHA with a version comment.**
`uses: actions/checkout@11bd719` plus `# v4.2.2` gives immutability and readability.
Tags move; SHAs do not.

**Do: key caches on lockfile content.**
`key: linux-go-${{ hashFiles('go.sum') }}` invalidates exactly when dependencies change.

**Do not: `npm install` in CI.** It resolves floating ranges at run time. `npm ci` installs
exactly the locked tree or fails loudly, which is the desired behavior.

**Do not: share one mega-workflow for all concerns.** A single 40-job workflow couples
lint flakes to deploy risk. Split into `ci.yaml` (verify), `security.yaml` (scan),
`release.yaml` (publish), each with its own triggers and approvers.

**Do not: `pull_request_target` with untrusted checkout and secrets.** That combination
executes attacker-controlled code with credential access. Keep untrusted code in
`pull_request` workflows with read-only tokens.

## 3. Workflow Configuration Example

```yaml
# .github/workflows/ci.yaml — staged, pinned, cache-safe Go pipeline
name: ci
on:
  pull_request:
  push:
    branches: [main]

permissions:
  contents: read

jobs:
  contract: # Stage 1 — under 90 seconds, fail fast
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@11bd719 # v4.2.2
      - uses: actions/setup-go@d35c59abb061a4a6fb18e82ac0862c1c55f65f549 # v5.5.0
        with:
          go-version: "1.24.3"
          cache-dependency-path: go.sum
      - run: gofmt -l . | grep . && exit 1 || true
      - run: go mod verify
      - uses: gitleaks/gitleaks-action@ff98106 # v2.3.9
  test: # Stage 2 — needs contract green
    needs: [contract]
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@11bd719 # v4.2.2
      - uses: actions/setup-go@d35c59abb061a4a6fb18e82ac0862c1c55f65f549 # v5.5.0
        with:
          go-version: "1.24.3"
          cache-dependency-path: go.sum
      - run: go test ./... -count=1
  scan: # Stage 3 — security audit
    needs: [test]
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@11bd719 # v4.2.2
      - run: go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
      - run: govulncheck ./...
```
