# W3C ARIA APG Keyboard Navigation Reference Guide

## 1. The Core Philosophy of Keyboard Accessibility
Mouse users explore spatial layouts directly by pointing and clicking. Keyboard users explore linearly (via `Tab`) and hierarchically (via composite widget keys). An accessible interface must balance efficiency (not requiring 50 Tab presses to reach a button) with discoverability.

## 2. Standard Keybindings by Widget Type

| Widget Type | Standard Keys | Action |
|---|---|---|
| **Button** | `Space`, `Enter` | Activate button action |
| **Link** | `Enter` | Navigate to URL |
| **Checkbox** | `Space` | Toggle checked state |
| **Tabs** | `ArrowLeft`, `ArrowRight` | Move focus and activate tab |
| **Tabs** | `Tab` | Exit tablist into selected `tabpanel` |
| **Menu / Menubar** | `ArrowDown`, `ArrowUp` | Move focus between menu items |
| **Menu / Menubar** | `Enter`, `Space` | Open submenu or activate item |
| **Menu / Menubar** | `Escape` | Close menu and return focus to toggle |
| **Dialog (Modal)** | `Tab`, `Shift+Tab` | Cycle through focusable elements inside dialog |
| **Dialog (Modal)** | `Escape` | Close modal and return focus to trigger |
| **Combobox** | `ArrowDown`, `ArrowUp` | Traverse suggestions |
| **Combobox** | `Enter` | Select suggestion and close listbox |

## 3. WCAG 2.2 Focus Appearance (SC 2.4.11) Rules
To satisfy the newest WCAG 2.2 AA Focus Appearance criteria:
- **Area**: The focus indicator must have an area at least as large as a 2px perimeter border around the unfocused component.
- **Contrast Ratio**: At least **3:1** contrast between the focused and unfocused states of the indicator pixels.
- **Not Obscured**: The item with focus must not be fully obscured by sticky banners or fixed footers.
