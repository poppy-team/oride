# Component Specification

## Purpose
Write implementation-ready component specifications: typed props API, variant and size matrices, controlled/uncontrolled state contracts, slot/event surfaces, and keyboard/ARIA behavior — so engineers build exactly what design intended with zero ambiguity.

## Use when
- Specifying a new or changed UI component (props, variants, states, events).
- Defining keyboard interaction, focus management, and ARIA contracts for a component.
- Enumerating the state matrix (default, loading, empty, error, disabled) and responsive breakpoints.
- Freezing a component API before implementation or design-system release.

## Do not use when
- Exploring open-ended UX directions (use `design-research`).
- Auditing whole screens for WCAG conformance (use `accessibility`).
- Theming tokens or global styles rather than per-component contracts (use `wireframe-styleguide`).

## Required context
- Component inventory: names, usage contexts, and design-token dependencies.
- Prop/state matrix requirements: variants, sizes, controlled vs uncontrolled states.
- Keyboard and screen-reader interaction contracts per component.

## Procedure
1. **Fix Identity & Scope**: Name the component, its single responsibility, and explicit non-goals. Example: `DataTable` owns tabular display, sorting, and row selection; it does not own pagination controls (separate `Paginator`).
2. **Type the Props API**: List every prop with name, type, required/optional, default, and constraints (e.g., `density: 'compact' | 'comfortable' = 'comfortable'`). Mark which props are controlled (parent-owned) vs uncontrolled (internal state with `default*` initializer).
3. **Enumerate Variants & States**: Build the variant × state matrix: variants (primary/secondary/danger), sizes (sm/md/lg), states (default, hover, focus, active, loading, empty, error, disabled). Every cell must have a defined visual and behavior; `N/A` cells need a reason.
4. **Contract Keyboard & ARIA**: Specify the full keyboard map (e.g., `ArrowUp/Down` move, `Home/End` jump, `Type-ahead` filters), focus trap/return rules for overlays, roving-tabindex vs aria-activedescendant choice, and required roles/attributes (`role="dialog"`, `aria-modal="true"`, labelledby/describedby wiring).
5. **Define Slots & Events**: List named slots (header, footer, empty-state) and emitted events with payloads (e.g., `onSortChange({ columnId, direction })`). Events carry data, never DOM nodes.
6. **Verify**: Run `scripts/verify.sh`. Confirm every prop is typed with defaults, every matrix cell resolved, keyboard map complete, and ARIA wiring explicit.

## Decision rules
- **No Untyped Props**: `any`-typed or undocumented props are forbidden; every prop has type, default, and constraint.
- **Controlled Clarity**: Each piece of state is declared controlled or uncontrolled — never both, never ambiguous.
- **Keyboard Completeness**: If a mouse user can do it, the spec defines the keyboard equivalent or explicitly justifies the exception.
- **Matrix Closure**: No `TBD` cells in the variant × state matrix at sign-off; unknowns become tracked follow-ups with owners.

## Evidence required
- Component spec sheet with typed props, variants, and state contracts.
- Keyboard/ARIA contract with focus-management rules.
- State-matrix fixtures covering default/loading/empty/error/disabled.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Component specification sheets with typed props, variants, and state contracts.
- Keyboard/ARIA contract per component with focus-management rules.
- State-matrix test fixtures covering default/loading/empty/error/disabled.

## Stop conditions
- All components specified with props, variants, and keyboard contracts plus state fixtures.
- Matrix fully closed; exceptions justified in writing.
- Token budget exhausted.

## Escalation rules
- Escalate to design-system lead if a component's scope overlaps another component's contract (ownership decision).
- Escalate immediately if accessibility requirements conflict with the requested visual design (needs product-level tradeoff).
