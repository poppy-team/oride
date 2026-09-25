# Bilingual contract (en-US canonical, pt-BR mandatory)

- en-US is written first; pt-BR is part of Done (pairs at CURRENT or
  IN_REVIEW, never silent STALE).
- Single source in `docs/shared/` for IDs, schemas, tokens, API/CLI
  identifiers, paths, code. Never duplicate or translate identifiers,
  commands, paths, URLs, semantic IDs, schema keys, enum/literal values.
- MUST/SHOULD/MAY map to DEVE/DEVERIA/PODE; keep one shared glossary.
- Canonical change pipeline: delta -> translate impacted pairs only ->
  parity review -> status CURRENT. Pair states: CURRENT, STALE, MISSING,
  IN_REVIEW, DRIFT_DETECTED, BLOCKED, NOT_APPLICABLE.
