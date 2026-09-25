# Wireframe Styleguide & Schematic System Specification

## Purpose
Define and enforce a standardized schematic design system for low-fidelity wireframes and structural prototypes. Establish graybox design tokens (neutral color ramps, wireframe typography, spacing scales, placeholder symbols), standard component state representations (default, hover/focus, active, disabled, error), and systematic annotation conventions to ensure clear, unambiguous communication between UX architecture and UI engineering before high-fidelity visual styling.

## Use when
- Establishing or updating the schematic visual language for wireframes and rapid prototypes.
- Defining graybox design tokens and structural component kits for UX and product planning.
- Documenting component interaction states (default, hover, focus, disabled, loading, invalid) at low fidelity.
- Standardizing wireframe annotation schemas, callout markers, and responsive layout behavior tables.
- Auditing low-fidelity deliverables for visual hierarchy consistency and accessibility baseline compliance.

## Do not use when
- Implementing production UI components or production design systems (use `ui-implementation` or `design-system`).
- Creating detailed visual design tokens with brand color palettes and production theming (use `design-tokens`).
- Formulating high-fidelity visual regression or pixel-perfection suites (use `visual-regression` or `visual-qa`).

## Required context
- Target platform constraints (Web, iOS, Android, Desktop) and target viewport ranges.
- Product functional requirements, user journeys, and information architecture hierarchy.
- Established or planned design token taxonomy (semantic names, spacing units).
- Accessibility baseline criteria (minimum 3:1 contrast for UI controls/boundaries, 4.5:1 for body text, 44x44px minimum target sizes).

## Procedure
1. **Establish Graybox Token Ramp**:
   - Define a restricted, functional neutral palette: Canvas (`#f8fafc`), Surface (`#ffffff`), Neutral Subdued (`#f1f5f9`), Border Subtle (`#e2e8f0`), Border Strong (`#768296`), Text Secondary (`#64748b`), Text Primary (`#0f172a`), Focus Ring (`#2563eb`), Error Accent (`#dc2626`).
   - Define structural typography tokens based on standard monospace or sans-serif families: Display, Heading 1-3, Body regular/medium, Caption, Code.
   - Enforce an 8pt base grid for structural spacing (`8px`, `16px`, `24px`, `32px`, `48px`, `64px`) and 4pt for micro-alignments (`4px`, `12px`).
2. **Standardize Schematic Symbols & Placeholders**:
   - Define standard wireframe iconography (bounding box with diagonal cross `X` for images, wavy horizontal lines for text blocks, chevron indicators for accordions/drawers).
   - Standardize avatar placeholders (circle with generic silhouette or initials), video placeholders (box with triangle play icon), and chart placeholders (bar/sparkline silhouettes).
3. **Specify Component State Matrices**:
   - For every interactive component (Button, Input, Dropdown, Toggle, Checkbox, Tab, Modal):
     - Detail 6 core states: *Default*, *Hover*, *Focus-Visible* (enforcing 2px high-contrast outline), *Active/Pressed*, *Disabled* (opacity reduction with distinct border pattern), *Invalid/Error* (inline indicator with distinct dashed or 2px outline and accessible error text).
     - Specify hit targets guaranteeing $\ge 44 \times 44\text{px}$ touch targets across mobile breakpoints.
4. **Define Systematic Annotation Conventions**:
   - Numbered callouts: Enclosed circles `[1]`, `[2]`, `[3]` linked to an adjacent tabular ledger.
   - Annotation ledger columns: ID, Component/Region, Interaction Trigger, System Behavior, Error Handling, Responsive Behavior.
   - Interaction flow indicators: Dashed vectors `-->` for state transitions, double-headed arrows `<-->` for bi-directional synchronizations.
5. **Formulate Responsive Layout Matrices**:
   - Document behavior across canonical breakpoints: Mobile (`320px - 767px`), Tablet (`768px - 1023px`), Desktop (`1024px - 1439px`), Wide (`1440px+`).
   - Document column counts (Mobile: 4-col, Tablet: 8-col, Desktop: 12-col), gutters (`16px`/`24px`), and margins (`16px`/`32px`).
   - Specify element reflow: stacking order, element collapse, side drawer conversion, and sticky landmark behaviors.
6. **Execute Verification & Handoff**:
   - Validate that all wireframe components strictly consume the graybox token ramp.
   - Run verification script `scripts/verify.sh` to confirm token compliance and annotation completeness.

## Decision rules
- **Color Discipline**: Never introduce arbitrary decorative colors into the wireframe styleguide. Only neutral tones (black, white, grays) are permitted, with isolated functional accents for keyboard focus (`#2563eb`) and form validation errors (`#dc2626`).
- **Typography Uniformity**: Use maximum 2 font families (one clean sans-serif for UI labels, one monospace for programmatic annotations/code).
- **Contrast Integrity**: All graybox text tokens must achieve $\ge 4.5:1$ contrast against their assigned background token. All UI borders and interactive boundaries must achieve $\ge 3:1$ contrast against adjacent surfaces.
- **Explicit Touch Targets**: Every button, input, tab, and navigation link must explicitly document bounding box dimensions $\ge 44 \times 44\text{px}$ on mobile layouts.
- **Zero Ambiguous Placeholders**: Placeholders must carry semantic descriptors (e.g., `[Image: 16:9 Hero Graphic]` rather than empty blank boxes).

## Evidence required
- Canonical wireframe styleguide document (`wireframe-styleguide-spec.md` or equivalent).
- Component state matrix covering all primary interactive controls across 6 standard states.
- Graybox token dictionary with hex codes, contrast ratios, and design token cross-references.
- Execution logs from `scripts/verify.sh` demonstrating zero token and annotation violations.

## Output contract
- Completed Wireframe Styleguide Specification adhering to `templates/wireframe-styleguide-spec.md`.
- Graybox CSS / Token declaration file (CSS Variables or JSON token schema).
- Component state reference diagram or markdown table.
- Verification pass log confirming structural and contrast compliance.

## Stop conditions
- Complete graybox token scale, component states, and annotation rules specified and verified.
- Passing run of `scripts/verify.sh`.
- Token budget exhausted or unresolvable conflict with product requirements.

## Escalation rules
- Escalate to Design Lead if product requires full color or brand aesthetics prior to wireframe sign-off (violates low-fidelity separation of concerns).
- Escalate to Accessibility Lead if proposed graybox contrast tokens fail WCAG 2.1 AA contrast requirements.
- Escalate to Technical Lead if requested component states require architectural state machines not supported by the runtime.
