# Wireframe Styleguide & Schematic System — Verification Checklist

## 1. Graybox Token Ramp & Contrast Verification
- [ ] Graybox color ramp is strictly restricted to neutral grays, with functional focus blue and error red.
- [ ] Canvas vs Surface contrast provides adequate visual separation (minimum $\Delta L^* \ge 5$).
- [ ] Text tokens meet WCAG 2.1 AA contrast requirements:
  - [ ] Body / Primary text (`--wire-text-primary`) achieves $\ge 4.5:1$ against surface/canvas.
  - [ ] Subdued / Secondary text (`--wire-text-secondary`) achieves $\ge 4.5:1$ against surface/canvas.
  - [ ] Component boundaries and structural borders (`--wire-border-strong`) achieve $\ge 3:1$ against adjacent backgrounds.
- [ ] Spacing tokens adhere strictly to the 8pt base grid (`8px`, `16px`, `24px`, `32px`, `48px`, `64px`) with 4pt for micro-adjustments.
- [ ] Typography scale is documented with explicit font sizes, line heights, font weights, and target uses.

## 2. Schematic Symbols & Placeholder Standards
- [ ] Image placeholders specify aspect ratio, bounding box style (diagonal `X`), and descriptive label.
- [ ] Media placeholders (video, audio, maps) have distinct, standardized schematic icons.
- [ ] Avatar placeholders provide circular bounding with initials or standard silhouette.
- [ ] Text blocks use structured placeholder formatting (wavy lines or labeled paragraph boxes) rather than raw lorem ipsum where readability is impacted.
- [ ] Data visualization placeholders document chart type (bar, line, donut) and axis landmarks.

## 3. Component State Matrices
- [ ] All primary interactive controls (Buttons, Inputs, Selects, Toggles, Checkboxes, Tabs, Dialogs) document 6 standard states:
  - [ ] **Default**: Rest state with standard border and neutral fill.
  - [ ] **Hover**: Slight elevation or background tint adjustment without color shift.
  - [ ] **Focus-Visible**: 2px solid high-contrast outline (`--wire-focus`) with 2px offset.
  - [ ] **Active / Pressed**: Darker shade or inset shadow indicating tactile press.
  - [ ] **Disabled**: Muted text, diagonal hatching or reduced opacity ($0.4$), `aria-disabled="true"`.
  - [ ] **Invalid / Error**: 2px error border (`--wire-error`), alert icon, and inline error message below control.
- [ ] Touch targets on mobile configurations guarantee minimum $44 \times 44\text{px}$ bounding box.

## 4. Annotation & Schematic Conventions
- [ ] Numbered callout badges (`[1]`, `[2]`, `[3]`) are placed adjacent to interactive components.
- [ ] Annotation ledger is fully populated with:
  - [ ] Callout Reference ID.
  - [ ] Component Name & Landmark Role.
  - [ ] User Trigger (Click, Tap, Keypress, Focus).
  - [ ] Expected System Action & State Change.
  - [ ] Error & Empty State Handling.
  - [ ] Responsive Reflow Behavior.
- [ ] State transition vectors use standardized arrows (`-->` forward transition, `<-->` bi-directional, `-.->` asynchronous background event).

## 5. Responsive Grid & Breakpoint Specifications
- [ ] Breakpoints are defined: Mobile (`320px`), Tablet (`768px`), Desktop (`1024px`), Wide (`1440px`).
- [ ] Grid column counts, gutter widths, and outer margins are specified for each breakpoint.
- [ ] Component reflow rules (stacking, collapsing, wrapping, off-canvas drawers) are explicitly annotated.

## 6. Engineering Alignment & Automated Verification
- [ ] Graybox tokens map 1:1 to planned production design tokens (`--wire-color-*` $\to$ `--color-*`).
- [ ] Styleguide deliverable matches `templates/wireframe-styleguide-spec.md`.
- [ ] Automated verification script `scripts/verify.sh` executes successfully with exit code 0.
