# UI/UX Contract (English projection)

This directory is the English projection of `docs/ui-ux/`, which is canonical.

**The keymap table is not duplicated here.** It is generated from
`internal/config/keybindings.go` into `docs/ui-ux/keymap.md` and guarded by
`TestDocumentedKeymapMatchesTheBindings`. A second copy in English would be a
second thing to drift, and a keymap that disagrees with the code is worse than no
keymap at all — see `docs/ui-ux/README.md` for the rule.

| Document | Canonical source | Status |
|---|---|---|
| Keymap | [`../ui-ux/keymap.md`](../ui-ux/keymap.md) | generated, guarded |
| Focus graph | [`../ui-ux/focus-graph.md`](../ui-ux/focus-graph.md) | Portuguese only for now |
| State vocabulary | [`../ui-ux/states.md`](../ui-ux/states.md) | Portuguese only for now |
| Layout | [`../ui-ux/layout.md`](../ui-ux/layout.md) | Portuguese only for now |
| Degradation | [`../ui-ux/degradation.md`](../ui-ux/degradation.md) | Portuguese only for now |

Translating a document that a test does not check adds a copy without adding a
guard. These are translated when their check exists, so the English version ships
with the same enforcement as the Portuguese one.
