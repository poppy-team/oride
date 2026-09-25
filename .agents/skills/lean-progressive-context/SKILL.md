# Lean Progressive Context

## Purpose
Select the smallest sufficient context for a task under strict token budgets through staged loading: start from a minimal working set, expand one layer at a time with a running token ledger, and stop at the first layer that answers the task.

## Use when
- Starting any Goal or Task where the relevant codebase exceeds what fits comfortably in the context window.
- Deciding what to load next (source files, specs, ADRs, test fixtures) instead of dumping whole directories.
- Defining or enforcing token budgets, expansion caps, and stopping conditions for agent runs.
- Writing Working Context Capsules that summarize completed stages for downstream agents.

## Do not use when
- Compressing or rewriting context that is already loaded (use `context-optimization`).
- Authoring prompts or eval harnesses (use `prompt-engineering`).
- Performing the domain work itself once context is sufficient; switch to the domain skill.

## Required context
- Task statement with explicit acceptance criteria.
- Total token budget for the run (default 60,000 tokens) and per-stage expansion cap (default 8,000 tokens).
- Entry-point hints: file paths, symbols, or specs most likely relevant.
- Stopping condition: what evidence proves the current layer is sufficient.

## Procedure
1. **Declare the budget up front**: record total budget (e.g. 60,000 tokens), per-stage cap (8,000 tokens), and max expansion stages (5) in the run ledger before loading anything.
2. **Load the minimal working set first**: task statement, entry-point file, and directly referenced spec only; measure with a tokenizer and log the running total.
3. **Expand one layer at a time**: each expansion names a hypothesis (e.g. "auth failure originates in token refresh, load auth/refresh.go next"), loads only that layer, and appends cost to the ledger.
4. **Test sufficiency after every layer**: ask whether acceptance criteria can now be addressed; if yes, stop expanding and start domain work.
5. **Write a capsule on stage completion**: a 10-line Working Context Capsule (goal, decisions, open questions, next layer if needed) so downstream agents inherit understanding without reloading raw context.
6. **Run `scripts/verify.sh`** to confirm ledger discipline and capsule presence before handoff.

## Decision rules
- **Budget first, loading second**: no file is read before its cost is chargeable against a declared budget.
- **One hypothesis per expansion**: each layer must state what it expects to find; fishing expeditions without hypotheses are prohibited.
- **Stop at sufficiency**: expanding past the layer that answers the task is a violation, not thoroughness.
- **Capsules over raw dumps**: handoffs carry capsules, never unbounded file lists.
- **Ledger honesty**: every loaded file and its token cost appears in the ledger; unlogged loads are treated as budget overruns.

## Evidence required
- Token ledger listing every loaded file with per-file and running-total costs.
- Working Context Capsule for each completed stage.
- Sufficiency note stating which layer answered the task and why expansion stopped.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Minimal sufficient working set actually used for the domain work.
- Token ledger proving spend stayed within budget and caps.
- Capsule chain enabling any reviewer to reconstruct the reasoning path.

## Stop conditions
- Acceptance criteria addressable with the current layer; expansion halts.
- Max expansion stages (default 5) reached without sufficiency.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to the task owner if sufficiency is unreachable within budget and scope reduction is needed.
- Escalate to a human if the entry-point hints are missing and blind search would burn the budget.
