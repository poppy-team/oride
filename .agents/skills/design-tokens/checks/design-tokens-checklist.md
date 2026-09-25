# W3C Design Tokens & Theming Invariants Checklist

## 1. DTCG Schema Conformance
- [ ] Every token definition includes explicit `$value` and valid `$type` attributes.
- [ ] Valid `$type` values: `color`, `dimension`, `duration`, `cubicBezier`, `number`, `fontFamily`, `fontWeight`, `shadow`, `typography`.
- [ ] Token aliases adhere to bracketed path notation (`"{color.neutral.500}"`).
- [ ] Token JSON parses without schema validation errors.

## 2. Three-Tier Architectural Hierarchy
- [ ] **Tier 1 (Primitives)**: Pure values without contextual meaning (e.g. `color.blue.600`, `spacing.4`).
- [ ] **Tier 2 (Semantics)**: Purpose-driven abstractions (e.g. `color.action.primary.default`, `color.surface.card`).
- [ ] **Tier 3 (Components)**: Scoped component overrides (e.g. `button.primary.bg`, `card.border`).
- [ ] **Rule Enforcement**: UI components never bind directly to Tier 1 primitives.

## 3. Theming & Contrast Invariants
- [ ] Light and dark themes defined by swapping Tier 2 semantic aliases.
- [ ] All semantic text-on-surface pairings satisfy WCAG 2.2 AA ($4.5:1$ normal text, $3.0:1$ large text).
- [ ] Interactive component boundaries and focus indicator tokens satisfy WCAG $\ge 3.0:1$ contrast against adjacent surfaces.
- [ ] High-contrast theme provided for accessibility conformance.

## 4. Spacing, Typography & Responsive Sizing
- [ ] Spacing tokens scale on a predictable 4px/8px modular grid using relative `rem` units.
- [ ] Typography tokens define font-size, line-height, and letter-spacing proportionally.
- [ ] No hardcoded physical pixel values in layout containers.

## 5. Implementation & Static Cleanliness
- [ ] Zero hardcoded hex colors (`#[0-9a-fA-F]{3,8}`) in component source files.
- [ ] Zero inline styles overriding design tokens without documented rationale.
- [ ] Automated token compilation succeeds for all target platforms (CSS, TS, Rust).
