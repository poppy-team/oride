# Keyboard Navigation & Operability Verification Checklist

## 1. Sequential Navigation & Tab Order
- [ ] Tab order strictly follows the logical visual reading sequence.
- [ ] Zero positive tabindexes (`tabindex="1"`, etc.) in the entire codebase.
- [ ] Skip navigation link provided on pages with repetitive header/nav structures.
- [ ] Hidden off-screen menus or drawers have `visibility: hidden` or `display: none` when closed to remove them from tab order.

## 2. ARIA APG Widget Patterns
- [ ] Composite widgets (menus, tabs, toolbars, listboxes) implement roving tabindex or `aria-activedescendant`.
- [ ] `ArrowUp` / `ArrowDown` or `ArrowLeft` / `ArrowRight` navigate within composite widget items.
- [ ] `Home` and `End` jump to the first and last elements in lists/toolbars.
- [ ] `Escape` dismisses menus, popovers, and dialogs, restoring focus to trigger.

## 3. Focus Visibility & Traps
- [ ] Focus indicators are clearly visible on all interactive elements (minimum 2px thickness).
- [ ] Focus indicator contrast is >= 3:1 against surrounding background and element colors.
- [ ] Zero blind `outline: none` without immediate `:focus-visible` replacement.
- [ ] Zero keyboard traps: all components can be exited via `Tab`, `Shift+Tab`, or `Escape`.
