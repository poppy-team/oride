# Editor Tooling Reference Guide

## Document Model

The document is the source of truth; widgets and viewport objects are projections. Every entity has a stable ID, schema version, and validated fields. View state such as selection, hover, tool mode, and panel layout lives in editor session state.

## Command Transactions

Represent every user-visible mutation as a command with `apply`, `revert`, label, affected IDs, and merge policy. Pointer drags coalesce into one command; failed operations roll back completely. Undo and redo form one bounded history with explicit memory accounting.

## Schema-Driven Inspectors

Generate controls from the same schema used by runtime validation. Custom editors register by type. Unknown future types open in a read-only inspector with schema and raw-value views rather than crashing or coercing data.

## Gizmos and Viewports

Convert pointer rays through the active camera, transform the hit point into the selected entity's local space, apply axis constraints, snapping, and fine modifiers, then execute one command. Multi-selection exposes a deterministic pivot and reports mixed values explicitly.

## Persistence and Recovery

Write documents through a validated, versioned serializer. Save to a temporary sibling, flush, and atomically replace the target. Migration must be deterministic and retain a recoverable prior revision.

## Patterns and Anti-Patterns

| Do | Avoid |
|---|---|
| Mutate the document through commands | Move the rendered object directly |
| Generate inspectors from runtime schema | Maintain a second hand-written field list |
| Keep editor state out of saves | Store selection in the document |
| Bound undo memory | Keep an unlimited command history |
| Test runtime symbol absence | Hide editor code with an unused source file |

## Short Example

Dragging a transform from `1.13` to `1.37` with 0.25-unit snapping records one `SetTransform` command with old value `1.0` and new value `1.5`; one undo restores the exact document hash.
