# Design System Quality & Engineering Checklist

## 1. Token Integration & Multi-Theming
- [ ] **Tiered Architecture**: Tokens follow Global -> Semantic -> Component hierarchy.
- [ ] **No Hardcoded Primitives**: Zero hardcoded hex colors, pixel margins/paddings, or raw font families in component styles.
- [ ] **Theme Switching**: Dark mode and high-contrast modes swap cleanly without layout shifts or page reloads.
- [ ] **Fallbacks**: CSS custom properties include valid fallbacks or are defined on `:root`.

## 2. Component API & Architecture
- [ ] **Polymorphism**: Supports `asChild` / Slot pattern to prevent invalid nested interactive elements.
- [ ] **Composition**: Complex widgets use compound components (`Menu.Root`, `Menu.Item`, `Menu.Trigger`) rather than oversized prop sets.
- [ ] **Zero External Margins**: Components do not declare outer margins (`margin-top`, `margin-bottom`); spacing is managed by layout containers.
- [ ] **Strict Typing**: All variants, sizes, and states are defined with TypeScript string literal unions or typed variant helpers.
- [ ] **Controlled & Uncontrolled**: Input/interactive components support both controlled (`value`, `onChange`) and uncontrolled (`defaultValue`) modes.

## 3. Accessibility & Keyboard Hygiene
- [ ] **Semantic Markup**: Uses native HTML elements (`<button>`, `<dialog>`, `<nav>`, `<input>`) where appropriate.
- [ ] **ARIA APG Compliance**: Complex widgets implement full keyboard patterns (Arrow navigation, Home/End, Esc, Enter, Space).
- [ ] **Focus Visible Indicator**: Clearly visible `:focus-visible` ring on all interactive states, conforming to SC 2.4.7 and SC 2.4.11.
- [ ] **Disabled State Contract**: Disabled elements communicate state via `aria-disabled="true"` or `disabled` and are removed from the sequential tab order appropriately.
- [ ] **Automated A11y Tests**: Axe-core scans return 0 violations across all variants and states.

## 4. Documentation & Packaging
- [ ] **Storybook Stories**: Interactive stories cover all variants, sizes, states (hover, active, focus, disabled, loading).
- [ ] **Live Controls**: Args are exposed with proper controls and TypeScript descriptions.
- [ ] **Tree-shaking**: Library package exports are ES modules with `"sideEffects": false` in `package.json`.
- [ ] **Deprecation Warnings**: Deprecated props log non-intrusive console warnings in development mode only.
