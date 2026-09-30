# Layout and Width Classes

## Classes

| Class | Width | What changes |
|---|---|---|
| `compact` | < 80 | The tree overlays instead of sitting side by side; the statusline shortens labels; the terminal panel takes the full width |
| `standard` | 80–119 | Full layout: tree, editor, and statusline side by side |
| `wide` | ≥ 120 | Like `standard`, with extra width for the tree and for the diagnostics column |

The limits are **inclusive at the bottom**: 80 is `standard`, 120 is `wide`.

## Rules

1. **Width is decided once per frame**, from the size the runtime
   hands over. No surface recomputes the class on its own.
2. **Nothing is cut off silently.** Whatever does not fit is truncated with a
   visible marker; a label that vanishes without a trace is indistinguishable from a label that never
   existed.
3. **Width is conservative.** An ambiguous-width character is treated as wide.
   Erring wide costs one column; erring narrow misaligns the entire
   line.
4. **Zero is a valid size.** A terminal that has not been measured yet must not make the
   editor panic or divide by zero.
