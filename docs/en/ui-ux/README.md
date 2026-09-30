# UI/UX Contract (English projection)

This directory is the English projection of `docs/ui-ux/`, which is canonical.

**The keymap table is generated, and the Portuguese file is the guarded one.** It is
generated from `internal/config/keybindings.go` into `docs/ui-ux/keymap.md` and guarded by
`TestDocumentedKeymapMatchesTheBindings`. [`keymap.md`](keymap.md) in this directory is an
English translation of that file (prose and headings only; the key/id pairs are
identical), but the test does not check it: when the bindings change, regenerate the
canonical table first and then carry the change over here — see `docs/ui-ux/README.md` for the rule.

| Document | Canonical source | Status |
|---|---|---|
| Keymap | [`../ui-ux/keymap.md`](../ui-ux/keymap.md) | translated in [`keymap.md`](keymap.md); canonical file generated, guarded |
| Focus graph | [`../ui-ux/focus-graph.md`](../ui-ux/focus-graph.md) | Portuguese only for now |
| State vocabulary | [`../ui-ux/states.md`](../ui-ux/states.md) | Portuguese only for now |
| Layout | [`../ui-ux/layout.md`](../ui-ux/layout.md) | translated in [`layout.md`](layout.md) |
| Degradation | [`../ui-ux/degradation.md`](../ui-ux/degradation.md) | Portuguese only for now |

Translating a document that a test does not check adds a copy without adding a
guard. The documents still marked "Portuguese only for now" are translated when
their check exists, so the English version ships with the same enforcement as the
Portuguese one.
