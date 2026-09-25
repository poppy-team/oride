---
name: design-system
description: Systematic UI engineering covering token architecture (DTCG), polymorphic accessible headless components, variant prop contracts, multi-brand theming, and regression testing gates.
---

# Design System Engineering

## 1. Purpose
Define, engineer, and govern a multi-platform, token-driven Design System that provides reusable, accessible, cohesive, and themeable UI components across the entire product ecosystem, ensuring zero visual drift and strict contract compliance.

---

## 2. Use When
- Creating or refactoring shared component libraries and UI design tokens.
- Establishing multi-brand, dark mode, or responsive theme architectures.
- Defining component API contracts (props, slots, variants, events, accessibility states).
- Implementing headless accessible primitives (Radix UI, React Aria, AccessKit) with custom styling.
- Establishing visual regression testing suites and interactive living documentation (Storybook, Styleguidist).

---

## 3. Do Not Use When
- Building one-off, highly specialized, domain-specific landing pages with zero reuse potential.
- Prototyping disposable mockups where formal token pipelines and semver contracts introduce unnecessary overhead.
- Auditing low-level memory layout or binary runtime internals (use `memory-management` or `clean-code`).

---

## 4. Required Context
Before implementing or modifying design system primitives, obtain:
1. **Design Tokens Specification**: DTCG-compliant JSON tokens covering colors, spacing, typography, radii, elevation, motion.
2. **Component Anatomy & States**: Visual specs and interaction matrices for all 19 states (default, hover, focus-visible, active, disabled, loading, error, etc.).
3. **Accessibility Baseline**: Required ARIA APG pattern, keyboard navigation schema, and contrast target (WCAG 2.2 AA / AAA).
4. **Target Platform & Framework**: Target runtimes (React, Vue, Svelte, Web Components, Flutter, Native).

---

## 5. Procedure

### Step 1: Token Integration & Theming Layer
1. Map raw design decisions into the 3-tier DTCG hierarchy:
   - **Global (Tier 1)**: Immutable raw values (`color.blue.500: #2563eb`).
   - **Semantic (Tier 2)**: Intent-driven abstraction (`color.action.primary.bg: {color.blue.500}`).
   - **Component-Scoped (Tier 3)**: Isolated component styling (`button.primary.bg: {color.action.primary.bg}`).
2. Expose tokens via CSS Custom Properties or platform-native variables under theme scopes (`:root`, `[data-theme="dark"]`, `[data-theme="high-contrast"]`).
3. Ensure runtime switching carries zero layout shift or JavaScript recalculation overhead.

### Step 2: Component Contract & Polymorphic Architecture
1. Define strict TypeScript interfaces for all component props using discriminated unions or variant maps (e.g., CVA / Tailwind Variants / StyleX).
2. Implement polymorphic rendering via the `asChild` / Slot pattern to prevent tag soup and support arbitrary wrappers without breaking semantic markup:
   ```tsx
   <Button asChild variant="primary">
     <a href="/login">Log In</a>
   </Button>
   ```
3. Separate component state/behavior from styling using headless primitives or well-defined hook contracts.

### Step 3: Accessibility & Keyboard Interaction Contract
1. Enforce native semantic elements wherever possible (`<button>`, `<dialog>`, `<nav>`).
2. If custom composite widgets are required (e.g., Combobox, Tabs, Menu), wire full ARIA APG state management, focus traps, and roving tabindexes.
3. Guarantee visible focus rings (`:focus-visible`) styled via semantic tokens (`outline: 2px solid var(--focus-ring); outline-offset: 2px`).

### Step 4: Component Documentation & Living Catalog
1. Create interactive Storybook stories (`*.stories.tsx`) covering every variant, size, and interactive state.
2. Expose interactive Controls (args) allowing consumers to test edge cases (long strings, right-to-left languages, zoom).
3. Document accessibility guarantees, keyboard shortcuts, and dos/don'ts.

### Step 5: Visual Regression & Contract Automated Gates
1. Run automated snapshot tests verifying rendered DOM structure.
2. Execute pixel-level visual regression tests (Playwright, Chromatic, Storybook Test Runner) across both light and dark themes.
3. Run automated accessibility scans (`axe-core`, `eslint-plugin-jsx-a11y`) in CI.

---

## 6. Decision Rules
1. **Composition Over Configuration**: Prefer compound components (`<Card.Header>`, `<Card.Body>`, `<Card.Footer>`) over monolithic prop-heavy components (`<Card title="" header="" subtitle="" footer="" ... />`).
2. **Zero Hardcoded Values**: Never write raw hex codes (`#1e293b`), pixel dimensions (`16px`), or font families directly in component stylesheets. Every value must resolve through design tokens.
3. **No Breaking Changes Without Codemods**: Renaming or removing a prop is a breaking change. Deprecate with runtime warnings and supply automated codemods before removal.
4. **Encapsulated Layout Boundaries**: Components must never specify external margins (`margin-top`, `margin-left`). Layout spacing is the responsibility of parent layout components (`Stack`, `Grid`, `Cluster`).

---

## 7. Evidence Required
- **Contract Tests**: TypeScript compilation with zero type errors under `strict: true`.
- **Accessibility Audit**: Automated Axe / Jest-Axe scan with 0 violations.
- **Visual Regression Proof**: Screenshot diffs confirming 0 unexpected pixel mutations across themes and breakpoints.
- **Storybook Coverage**: Stories present for 100% of defined component variants.

---

## 8. Output Contract
A production design system artifact must contain:
1. Component implementation file (e.g. `Button.tsx`, `Button.vue`).
2. Component style definition consuming semantic design tokens.
3. Component unit and accessibility tests (`*.test.tsx`).
4. Storybook documentation story (`*.stories.tsx`).
5. Export entry in package index with clean tree-shaking metadata (`package.json` with `"sideEffects": false`).

---

## 9. Stop Conditions
- All component variants pass type checking, unit tests, and accessibility validations.
- Visual regression suite matches baseline with zero unwanted diffs.
- Component API documentation is generated and verified in Storybook.

---

## 10. Escalation Rules
- Escalate to Design Lead if a proposed component variant cannot satisfy WCAG 2.2 AA contrast with the current color palette.
- Escalate to Architecture Board if a new component introduces a conflicting dependency or breaking token tier change.
