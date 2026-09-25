# Responsive Reflow, Text Resize & 400% Zoom Accessibility (WCAG 2.2 SC 1.4.4 & 1.4.10)

## 1. Purpose
Design, implement, and verify fluid user interfaces that support full responsiveness, user font scaling, and multi-directional reflow without horizontal scrolling or loss of functionality. This skill mandates strict adherence to **WCAG 2.2 SC 1.4.4 (Resize Text up to 200%) and SC 1.4.10 (Reflow at 400% zoom / 320 CSS pixels)**, relative typography sizing (`rem`/`em`), defensive container sizing, and elimination of fixed-pixel layout clipping.

---

## 2. Use When
- Building responsive web applications, desktop application webviews, or documentation portals.
- Testing layouts against 400% zoom on a standard 1280px display (equivalent to a 320px viewport).
- Verifying that browser default font adjustments (up to 200%) resize all textual content without truncation or overlap.
- Designing modals, navigation sidebars, data tables, and forms for constrained mobile viewports.
- Operating in mode(s): `testing`, `implementation`, `review`, `audit`.

---

## 3. Do Not Use When
- Developing native terminal UI (TUI) interfaces with fixed character cell grids (use `tui` protocol).
- Building specialized two-dimensional canvases (e.g. 3D viewports, spreadsheet grids, CAD drawing canvases) explicitly exempt under SC 1.4.10.
- Creating native game engine HUDs with fixed pixel coordinate buffers.

---

## 4. Required Context
Before auditing or implementing reflow behavior, verify:
- **Reflow Viewport Target**: 320 CSS pixels wide by 256 CSS pixels high (simulating 400% zoom on a 1280x1024 screen).
- **Text Scaling Target**: 200% browser font size increase without assistive technology screen magnifiers.
- **Exempt Content Classification**: Identify two-dimensional essential content (data tables, code diff blocks, schematic diagrams, canvas maps).
- **Styling Paradigm**: CSS Grid, Flexbox with `flex-wrap: wrap`, CSS `clamp()`, and relative container queries (`@container`).

---

## 5. Procedure

```
[UI Layout / Component CSS]
         |
         v
[1. Relative Units Enforcement] -----> Convert px fonts/spacing to rem/em
         |
         v
[2. Viewport 320px Reflow Check] ----> Flex wrap, column collapse, zero horizontal scroll
         |
         v
[3. Fixed Height & Overflow Audit] --> Eliminate fixed height + overflow:hidden on text
         |
         v
[4. Sticky / Fixed Header Budget] ---> Cap sticky chrome to < 30% of viewport height
         |
         v
[5. Automated Reflow Gauntlet] ------> Headless Playwright 320x256 snapshot verification
```

### Step 1: Relative Units & Font Scaling (SC 1.4.4)
1. Set root font size to browser default (`html { font-size: 100%; }` or `16px` default). Never lock root font size with an absolute pixel value like `html { font-size: 16px !important; }`.
2. Declare typography and container padding in relative units (`rem`, `em`, `ch`):
   ```css
   .card-title {
     font-size: 1.25rem; /* Scales when browser font size changes */
     line-height: 1.4;
   }
   ```
3. Use media queries declared in `em` units so breakpoints adjust when the user zooms in text:
   ```css
   @media (max-width: 48em) { /* 768px at 16px font; expands at 200% font */
     .layout-columns { flex-direction: column; }
   }
   ```

### Step 2: 400% Zoom / 320 CSS Pixels Reflow (SC 1.4.10)
1. Multi-column layouts must automatically collapse into a single vertical stream at widths $\le 320\text{px}$:
   ```css
   .grid-container {
     display: grid;
     grid-template-columns: repeat(auto-fit, minmax(min(100%, 18rem), 1fr));
     gap: 1rem;
   }
   ```
2. Modals, sidebars, and dialogue cards must use `max-width: 100%` or `max-width: calc(100vw - 2rem)`. Never specify fixed widths exceeding 300px without a fluid max-width clamp.

### Step 3: Fixed Heights & Clipping Eradication
- Never combine fixed heights (`height: 40px`, `max-height: 200px`) with `overflow: hidden` on elements containing user-facing text.
- Use `min-height` instead of `height` to allow containers to expand vertically when text scales to 200%:
  ```css
  /* FORBIDDEN: Text truncates on 200% zoom */
  .button { height: 40px; overflow: hidden; }

  /* COMPLIANT: Container expands vertically if text wraps */
  .button { min-height: 2.5rem; padding: 0.5rem 1rem; }
  ```

### Step 4: Sticky Chrome & Screen Real Estate Budget
At 400% zoom, the effective viewport height is often only $256\text{px}$:
- Sticky top navigation bars and bottom action bars must not exceed 25% of viewport height combined.
- Use `@media (max-height: 25em)` to un-stick fixed headers or convert them into static scrollable elements:
  ```css
  @media (max-height: 25em) {
    .site-header {
      position: static; /* Un-stick header on vertically constrained zoomed screens */
    }
  }
  ```

### Step 5: Handling Essential 2D Content
For tables or diagrams that require 2D scrolling:
- Wrap the table in an explicit scroll container with `overflow-x: auto` and `tabindex="0"`.
- Do not let the outer document window scroll horizontally. The horizontal scrollbar must be strictly confined to the table container.

---

## 6. Decision Rules & Invariants
- **RULE 1 (Zero Two-Dimensional Page Scrolling)**: The main document `window` must never exhibit horizontal scrolling at a 320px viewport width (except for explicitly isolated 2D data tables or maps).
- **RULE 2 (No Fixed Font Pixels)**: Typography declared with fixed `px` sizes that fail to respond to browser default font adjustments fails review immediately.
- **RULE 3 (No Text Truncation with overflow:hidden)**: Wrapping text in fixed-height containers with `overflow: hidden` or `text-overflow: ellipsis` that truncates critical content on text zoom is prohibited.
- **RULE 4 (Modal Fluidity)**: Modals and flyout menus must fit within a 320px viewport without clipping action buttons or escape controls.

---

## 7. Evidence Required
- **Headless Viewport Evidence**: Playwright automated test screenshots at $320\text{px} \times 256\text{px}$ proving absence of document horizontal scrollbar (`window.innerWidth === document.documentElement.scrollWidth`).
- **Text Scaling Evidence**: Screenshot verification showing all text containers expanding cleanly at 200% font scaling without overlap.
- **Reflow Audit Scorecard**: Full evaluation of all primary user workflows under 400% zoom.

---

## 8. Output Contract
- Responsive, fluid stylesheets conforming to SC 1.4.4 and SC 1.4.10.
- Zoom & Reflow Audit Report (`templates/zoom-reflow-report.md`).
- Automated Playwright/Cypress reflow test specifications.

---

## 9. Stop Conditions
- Document `scrollWidth <= innerWidth` verified at 320px viewport.
- All text readable with zero overlap or clipping at 200% font resize.
- Token budget exhausted.
- Blocked on external dependency.

---

## 10. Escalation Rules
- Escalate to UX lead if complex data visualization cannot be presented in 320px width without complete structural re-architecting into card lists.
- Escalate if third-party embedded iframe cannot support fluid responsive resizing.
