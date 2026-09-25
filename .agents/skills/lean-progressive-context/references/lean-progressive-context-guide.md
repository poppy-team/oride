# Lean Progressive Context — Reference Guide

## 1. Core Concepts

### 1.1 The Budget Envelope
Three numbers govern every run: total budget B (default 60,000 tokens), per-stage cap S
(default 8,000 tokens), and max stages N (default 5). The invariant is simple:
sum of stage costs stays under B, each stage stays under S, and stage count stays under
N. Breaching any of the three requires a written override, which makes budget creep
visible instead of silent.

### 1.2 Staged Expansion Model
Total cost decomposes per layer:

```
C_total = C_stage1 + C_stage2 + ... + C_stageK,  K <= N
sufficiency(K) = acceptance criteria addressable at layer K
```

The optimal stop is the smallest K with sufficiency true. Loading layer K+1 after
sufficiency at K is waste, and the ledger exists to make that waste auditable.

### 1.3 Working Context Capsules
A capsule is a 10 to 15 line handoff note: goal in one line, decisions taken, evidence
seen, open questions, and the hypothesized next layer. Capsules compress a stage's
8000-token working set into roughly 200 tokens, a 40x reduction that lets multi-stage
runs stay inside the envelope.

### 1.4 Sufficiency Testing
After each layer, answer one question: can the acceptance criteria now be addressed?
"Addressed" means the agent can name the files to change and the verification to run,
not that it has read every related file. Vague unease ("maybe I should check the whole
tree") is not insufficiency; a named missing fact is.

## 2. Patterns and Anti-Patterns

| Pattern (do this) | Anti-Pattern (never do this) |
|---|---|
| Declare B/S/N in writing before loading | Start reading, decide budget "later" |
| Load entry point + spec first (~3000 tokens) | Glob the whole repo tree into layer 1 |
| One hypothesis per expansion layer | "Load everything about auth just in case" |
| Stop at the first sufficient layer | Keep expanding after sufficiency "for context" |
| Capsule handoffs between stages | Forward 40,000 tokens of raw files downstream |
| Logged, measured token costs | Guessed costs, unlogged reads |

## 3. Worked Example
Task: fix token-refresh race in auth service, budget 60,000/8,000/5. Layer 1 loads the
bug report, auth/refresh.go, and the auth spec: 3,100 tokens. Hypothesis for layer 2:
"race is between refresh and revoke", loads auth/revoke.go plus two tests: 2,400 tokens,
running total 5,500. Sufficiency holds (files to change and test command identified),
expansion stops at K=2, capsule written in 11 lines, and 54,500 tokens remain unspent.
