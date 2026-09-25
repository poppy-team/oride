# Modes: GENESIS / ADOPTION / DELTA

## GENESIS (new projects)
Intent -> profiles -> manifest -> least workforce -> canonical en-US ->
pt-BR -> parity -> audit -> handoff. Never stamp implementation-ready while
structural decisions stay implicit.

## ADOPTION (existing projects)
1. Inspect repo, code, tests, docs, CI, config, runtime.
2. Gap Matrix per area: COMPLIANT / PARTIALLY_COMPLIANT /
   FUNCTIONAL_BUT_DIFFERENT / RUDIMENTARY / STUB / BROKEN / DUPLICATED /
   MISSING / OBSOLETE / UNKNOWN, each with evidence.
3. Preserve what works; fix only the delta. Restructuring needs strong
   evidence plus an ADR.

## DELTA (documented projects)
Classify the change, compute the impacted set (capabilities, contracts,
tests, docs, risks, migrations, translations), update only that set, emit
a delta record. Never rewrite or retraduce unaffected files.
