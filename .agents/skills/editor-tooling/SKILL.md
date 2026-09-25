# Editor Tooling & UI

## Purpose
Build deterministic, undo-safe desktop editor tooling: scene/document inspectors, viewport gizmos and overlays, multi-viewport coordination, command-based undo/redo, and stable serialization — all within frame budget and without leaking editor state into runtime builds.

## Use when
- Implementing scene inspectors, property editors, asset browsers, or outliner trees.
- Adding viewport gizmos (translate/rotate/scale), snapping, overlays, or multi-viewport sync.
- Designing undo/redo stacks, selection models, or document serialization for editor files.
- Auditing editor frame budget, interaction latency, or runtime/editor code separation.

## Do not use when
- Building end-user application UI rather than developer-facing editors (use `component-specification`).
- Implementing the runtime simulation itself (use `game-runtime`, `game-engine-architecture`).
- Authoring editor visual themes rather than editor behavior and data flow.

## Required context
- Editor surface inventory: inspectors, viewports, gizmos, and asset browsers in scope.
- Scene/document model schemas and undo-redo granularity requirements.
- Frame-budget and interaction-latency ceilings for editor operations.

## Procedure
1. **Model the Document**: Define the scene/document schema with stable IDs for every entity. All editor operations mutate this model — never the rendered view directly. Example: moving an object edits `transform.position` on entity `e-1042`; the viewport re-renders from the model.
2. **Command-ify Every Mutation**: Wrap each user action in a command object (`do`/`undo`, label, merge policy). Drag operations coalesce into one undo step on pointer-up; text edits commit per pause-group. The undo stack is bounded (e.g., 200 entries) with memory accounting.
3. **Build Inspectors from Schema**: Generate property editors from the same schema the runtime validates, so the editor cannot author invalid documents. Custom editors are registered per type; unknown types fall back to a read-only raw view, never a crash.
4. **Gizmos with Snapping Math**: Implement translate/rotate/scale gizmos in viewport space with explicit snapping (e.g., 0.25 m, 15°), axis locking via modifier keys, and numeric-entry override. Overlay rendering uses depth-tested lines with an on-top pass for the active handle only.
5. **Isolate Editor from Runtime**: Editor-only code compiles behind a feature flag and never ships in runtime builds; verify by symbol scan. Selection, hover, and overlay state live in editor session objects, never in the document model.
6. **Verify**: Run `scripts/verify.sh`. Prove undo/redo round-trips (100 random ops, model hash equal), serialization determinism (save→load→save byte-stable), and gizmo drags land on snapped values within interaction-latency budget.

## Decision rules
- **Model First**: No direct view mutation; every change flows through a command into the document model.
- **Undo Is Total**: Any user-visible mutation without an undo path is a defect — including selection-driven and bulk operations.
- **Schema Single Source**: Inspector widgets and runtime validators read the same schema; divergence is a release blocker.
- **Editor Never Ships**: Runtime binaries contain zero editor symbols; verified by automated scan per build.

## Evidence required
- Undo/redo fuzz log: random-op sequences with model-hash equality proof.
- Serialization determinism log: save→load→save byte-stability.
- Gizmo precision and interaction-latency measurements vs budget.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Editor tooling implementation with inspector/gizmo/viewport behaviors.
- Undo-redo and serialization conformance with deterministic fixtures.
- Interaction-latency benchmark evidence within frame budget.

## Stop conditions
- Editor interactions deterministic and within budget with undo-safe serialization.
- Editor/runtime separation verified by symbol scan.
- Token budget exhausted.

## Escalation rules
- Escalate to lead architect if undo granularity conflicts with collaborative/multi-user editing requirements (needs CRDT/OT decision).
- Escalate immediately if editor actions can corrupt documents without recovery path.
