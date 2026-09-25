# Responsive Reflow & Text Resize Invariants Checklist (WCAG 2.2)

## 1. 320 CSS Pixel Reflow Invariants (SC 1.4.10)
- [ ] Document does not require horizontal scrolling at a 320px viewport width (`scrollWidth === innerWidth`).
- [ ] Multi-column layouts collapse gracefully into a single vertical reading order.
- [ ] Modals, dialogs, and popovers fit within 320px with sufficient padding and accessible close buttons.
- [ ] Essential 2D scrollable content (e.g. wide data tables) is isolated within dedicated scroll containers with keyboard focusable handles (`tabindex="0"`).

## 2. Text Resizing up to 200% (SC 1.4.4)
- [ ] Typography declared using relative units (`rem`, `em`, `ch`), allowing scaling when browser default font size changes.
- [ ] Zero text truncation or ellipsis (`text-overflow: ellipsis`) on critical instructions or interactive labels.
- [ ] Containers holding text avoid fixed pixel heights (`height: 32px`) and use `min-height` with fluid vertical expansion.
- [ ] Adjacent text elements do not collide, overlap, or get clipped by sibling boundaries at 200% zoom.

## 3. Viewport Real Estate & Sticky Chrome
- [ ] Sticky headers and footers consume $< 25\%$ of viewport height at constrained heights ($256\text{px}$).
- [ ] Media query `@media (max-height: 25em)` un-sticks fixed headers on vertically constrained displays.
- [ ] Floating action buttons and chat widgets do not obscure primary form inputs or submit actions.

## 4. CSS Architecture & Unit Hygiene
- [ ] Zero usage of fixed `width: > 300px` without accompanying `max-width: 100%`.
- [ ] Media query breakpoints defined in relative `em` units to scale in harmony with user zoom levels.
- [ ] Fluid typography configured with `clamp(min, preferred, max)` where applicable.
