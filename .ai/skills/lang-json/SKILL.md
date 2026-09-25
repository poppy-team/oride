---
name: lang-json
description: JSON RFC 8259 compliance, JSON Schema contract validation, safe numeric bounds, canonical serialization, and zero comments.
---

# JSON Engineering & Schema Contract

## 1. RFC 8259 Compliance & Syntax Discipline
- Strict compliance with RFC 8259: zero comments (\`//\` or \`/* */\`), zero trailing commas, and double quotes for all strings and keys.
- Strictly reject duplicate object keys.

## 2. Schema-First Validation
- Machine contracts must be governed by a canonical JSON Schema (\`schemas/*.schema.json\`) referencing draft 2020-12 or draft 7.
- Schemas must declare \`additionalProperties: false\` on closed object models to prevent schema poisoning.

## 3. Numeric Precision & Security Limits
- Large integers exceeding 2^53 - 1 (JavaScript safe integer limit) must be represented as strings to prevent truncation.
- Enforce nesting depth limits (max 32 levels) and maximum payload byte sizes to prevent stack exhaustion attacks.
- Utilize deterministic / canonical key sorting when computing SHA-256 digests or signing payloads.\n