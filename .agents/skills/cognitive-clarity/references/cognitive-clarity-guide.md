# Cognitive Clarity & Neurodivergent Communication — Reference Guide

## 1. Core Concepts

### 1.1 Response-First Structure (TL;DR Lead)
Every technical response opens with the outcome in one or two sentences — what was done, what is blocked, or what decision is needed — before any background. Readers under high cognitive load (ADHD, dyslexia, fatigue) extract the payload even if they read nothing else. Detail follows in progressively disclosed sections, never before the lead.

### 1.2 Single-Decision Focus
Working memory holds roughly one open decision comfortably. Present the single most important fork with a recommended default and a compact tradeoff table (2–3 rows), instead of broadcasting five open questions. Secondary decisions queue explicitly as "parked" items with owners, so nothing is lost and nothing competes for attention.

### 1.3 Rationale Before Mechanism ("Why" → "How")
State the engineering purpose before the code diff. A reader who grasps *why* (e.g., "eliminate the TOCTOU window between check and open") can reconstruct or review the *how* with far less effort than one handed mechanics first. This ordering also exposes wrong-track work earlier.

### 1.4 Visual Anchoring and Chunking
Walls of text above four sentences collapse scanning. Use headings, numbered steps, checklists, and tables; bold the load-bearing terms so the eye can re-anchor after interruption. One instruction per step, one idea per bullet.

## 2. Patterns and Anti-Patterns

| Pattern (do) | Anti-pattern (avoid) |
|---|---|
| Lead with status: "Done: X verified. Next: Y." | Bury the outcome after three paragraphs of context |
| One targeted question with 2–3 concrete options | Open-ended "what do you think?" questionnaires |
| Numbered, decoupled verification steps with expected outputs | "Just run the tests and see" without oracle |
| Name assumptions explicitly ("Assuming Postgres 15…") | Silent assumptions discovered three turns later |
| Bounded output; link to details instead of pasting | Full-file dumps when a 10-line excerpt suffices |
| Blame-free diagnostics ("the check fails because…") | Normative language ("you should have…") |

## 3. Worked Example

Vague: "The build is kind of broken in a few places, maybe dependencies or config, can you look when you have time?"

Clear:
> **Status:** `billing-service` build fails; checkout flow untested. **Next:** pin `fastify` to 4.28 (CVE fix breaks 4.21 API).
> 1. Apply patch `billing/fastify-pin.diff` (2 lines).
> 2. Run `go test ./internal/billing/...` — expect 47/47 green.
> 3. If step 2 fails, paste only the first failing assertion (not the full log).
> Parked: Q4 migration of `expired-cache` (owner: backend-team, not blocking).

The rewrite states outcome first, isolates one decision, gives verifiable steps with oracles, bounds the log paste, and parks the non-urgent item instead of dropping it.
