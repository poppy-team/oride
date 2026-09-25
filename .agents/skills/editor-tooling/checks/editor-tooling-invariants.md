# Editor Tooling Verification Checklist

## Document and Command Model
- [ ] Stable IDs, schema versions, ownership, and validation protect the document.
- [ ] Every visible mutation executes through an apply/revert command.
- [ ] Drag, text, and bulk operations define explicit merge and undo behavior.

## Undo and Persistence
- [ ] Undo history is bounded, reports memory use, and restores exact model hashes.
- [ ] Failed or cancelled commands roll back without partial state.
- [ ] Save, load, migration, and atomic replacement are deterministic and recoverable.

## Inspectors and Viewports
- [ ] Inspectors and runtime validation use one schema source.
- [ ] Unknown future types open a read-only fallback instead of crashing.
- [ ] Gizmos apply camera, local-space, axis, snapping, and pivot math deterministically.

## Interaction and Separation
- [ ] Pointer drags complete within the interaction-latency budget and preserve focus state.
- [ ] Selection, hover, tool mode, and panel layout remain editor session state.
- [ ] Runtime builds contain no editor-only symbols or editor asset dependencies.

## Evidence
- [ ] Random operation sequences round-trip with identical document hashes.
- [ ] Save-load-save and migration fixtures are byte-stable where the format requires it.
- [ ] `scripts/verify.sh` exits with code 0 from the repository root.
