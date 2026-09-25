---
name: keyboard-accessibility
description: WCAG 2.2 keyboard navigation, ARIA APG interaction patterns, roving tabindex in composite widgets, visible focus indicators, skip-to-content links, and zero keyboard traps across Web, Native, and TUI.
---

# Keyboard Navigation & Operability Engineering

## 1. Title and Description
**Keyboard Navigation & Operability Engineering (`keyboard-accessibility`)**
Enforces complete, mouse-independent keyboard operability across Web applications, Desktop native software (egui, Slint), and Terminal User Interfaces (TUI). Implements W3C WAI ARIA Authoring Practices Guide (APG) keyboard patterns, roving tabindexes, clear focus indicators, and WCAG 2.2 Success Criteria.

## 2. Purpose
Guarantee that all interactive digital experiences can be operated quickly, intuitively, and without barriers by keyboard-only users, switch-device operators, and screen reader navigators, preventing inaccessible mouse-only interfaces and keyboard traps.

## 3. Prerequisites
- UI component markup or interface source code.
- Functional knowledge of ARIA Authoring Practices Guide (APG) design patterns.

## 4. Inputs
- View layout, interactive components (dialogs, menus, tables, tabs, forms).
- Target platform: Web (DOM), Desktop (AccessKit, OS UI automation), or TUI (Terminal).

## 5. Outputs
- Fully keyboard-operable component logic with zero mouse dependence.
- Automated keyboard navigation tests (Tab order, arrow key navigation, Escape handling).
- Audit report conforming to `templates/keyboard-accessibility-report.md`.

## 6. Execution Steps

### Step 1: Establish Logical DOM & Tab Order
1. **Reading Order Equals Tab Order**: The sequence of interactive elements in the DOM must strictly reflect the logical visual reading flow. Never reorder visual elements with CSS `order` or `flex-direction: reverse` without synchronizing the underlying DOM sequence.
2. **Tabindex Discipline**:
   - `tabindex="0"`: Inserts a naturally non-focusable element into the keyboard tab order.
   - `tabindex="-1"`: Removes an element from sequential tab order while retaining programmatic focusability (`element.focus()`).
   - **Strictly Forbidden**: `tabindex > 0` (positive tabindexes destroy predictable tab order and are prohibited).

### Step 2: Implement ARIA APG Composite Widget Patterns
Composite widgets (menus, tab lists, toolbars, grids, tree views) must present a single tab stop to the outer page, managing internal item navigation via arrow keys:
1. **Roving Tabindex Pattern**:
   - The currently active/focused item in the widget has `tabindex="0"`.
   - All sibling items in the widget have `tabindex="-1"`.
   - When the user presses `ArrowRight` or `ArrowDown`, change the current item to `tabindex="-1"`, set the next item to `tabindex="0"`, and call `.focus()`.
2. **Keyboard Mapping by Component Type**:
   - **Tabs**: `Left`/`Right` arrows switch active tab; `Tab` key moves focus directly into the active `tabpanel`.
   - **Dropdown Menus & Comboboxes**: `DownArrow` opens and navigates items; `Enter` selects; `Escape` closes and returns focus to the trigger button.
   - **Dialogs / Modals**: Immediately trap focus inside the modal; `Escape` closes the modal.
   - **Sliders**: `Left`/`Down` decreases value; `Right`/`Up` increases value; `Home`/`End` sets min/max.

### Step 3: Enforce Visible Focus Indicators (WCAG 2.2 SC 2.4.7 & 2.4.11)
1. **Never Remove Outlines Without Replacement**: Never author `outline: none` or `outline: 0` without immediately defining an alternative high-visibility `:focus-visible` style.
2. **Focus Appearance Standards**:
   - Contrast ratio of the focus indicator must be at least **3:1** against the background and against the unfocused component border.
   - Minimum thickness: **2px solid**, with an offset (e.g. `outline-offset: 2px`) to ensure the indicator is not clipped by parent containers.

### Step 4: Bypass Blocks & Skip Links (WCAG SC 2.4.1)
Provide a hidden-until-focused skip link as the very first interactive element on any page with global navigation:
```html
<a href="#main-content" class="sr-only focus:not-sr-only focus:absolute focus:p-4 focus:bg-white focus:text-black focus:z-50">
  Skip to main content
</a>
```

### Step 5: Prevent and Eliminate Keyboard Traps (WCAG SC 2.1.2)
1. Ensure the user can navigate into and out of any component using only standard keyboard controls (`Tab`, `Shift+Tab`, `Escape`).
2. If focus enters an embedded editor, iframe, or custom canvas, provide an explicit documented escape chord (e.g. `Esc` followed by `Tab`).

### Step 6: Native Desktop & TUI Keyboard Parity
1. **Desktop (egui / Slint / AccessKit)**: Ensure custom widgets implement keyboard event listeners (`egui::Event::Key`) and register focus nodes.
2. **Terminal (Bubble Tea)**: Support standard navigation keys (`Tab`, `Shift+Tab`, `Enter`, `Esc`) and optional vim keys (`j`/`k`) configurable via KeyMaps.

## 7. Verification
```bash
# 1. Automated keyboard navigation test
npm test -- src/tests/keyboard-navigation.test.ts

# 2. Audit script for positive tabindexes and outline suppression
./scripts/verify.sh .
```

## 8. Fallbacks & Failure Recovery
| Failure Class | Root Cause | Deterministic Remediation |
|---|---|---|
| Focus Lost on DOM Removal | An active focused element is deleted or unmounted | Before removing element from DOM, programmatically move focus to a logical parent or previous sibling (`previousSibling.focus()`). |
| Keyboard Trap in Modal | Modal does not handle Tab key wrapping | Use `focus-trap` library or native `<dialog>` element with `.showModal()`. |
| Focus Indicator Clipped | Parent container has `overflow: hidden` | Use `box-shadow` or `outline-offset: -2px` (inset indicator) instead of outer offset. |
| Arrow Keys Scroll Entire Page | Event listener fails to prevent default browser behavior | Call `event.preventDefault()` inside the specific `ArrowUp` / `ArrowDown` key handlers. |

## 9. Constraints
- **Zero positive tabindexes**: `tabindex="1"` or higher is strictly forbidden.
- **Complete parity**: Every single action performable by mouse click must have a documented, functional keyboard path.
- **Visible focus non-negotiable**: Focus rings must remain distinctly visible in all color themes and high-contrast modes.

## 10. Examples

### Anti-Pattern: Clickable Div Without Keyboard Support
```html
<!-- BAD: Mouse-only div, not in tab order, no keyboard activation, outline suppressed -->
<div onclick="openMenu()" style="outline: none; cursor: pointer;">
  Options Menu
</div>
```

### Idiomatic Pattern: ARIA APG Roving Tabindex Toolbar
```tsx
// GOOD: Full roving tabindex keyboard navigation, semantic roles, focus indicators
import React, { useRef, useState } from 'react';

interface ToolbarProps {
  items: Array<{ id: string; label: string; onAction: () => void }>;
}

export const AccessibleToolbar: React.FC<ToolbarProps> = ({ items }) => {
  const [focusedIndex, setFocusedIndex] = useState(0);
  const buttonRefs = useRef<(HTMLButtonElement | null)[]>([]);

  const handleKeyDown = (e: React.KeyboardEvent, index: number) => {
    let nextIndex = index;

    if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
      e.preventDefault();
      nextIndex = (index + 1) % items.length;
    } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
      e.preventDefault();
      nextIndex = (index - 1 + items.length) % items.length;
    } else if (e.key === 'Home') {
      e.preventDefault();
      nextIndex = 0;
    } else if (e.key === 'End') {
      e.preventDefault();
      nextIndex = items.length - 1;
    }

    if (nextIndex !== index) {
      setFocusedIndex(nextIndex);
      buttonRefs.current[nextIndex]?.focus();
    }
  };

  return (
    <div role="toolbar" aria-label="Editor Actions" className="flex gap-2 p-2 border rounded">
      {items.map((item, index) => (
        <button
          key={item.id}
          ref={(el) => (buttonRefs.current[index] = el)}
          tabIndex={focusedIndex === index ? 0 : -1} // Roving tabindex
          onClick={item.onAction}
          onKeyDown={(e) => handleKeyDown(e, index)}
          className="px-3 py-1.5 rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-600 focus-visible:ring-offset-2"
        >
          {item.label}
        </button>
      ))}
    </div>
  );
};
```
