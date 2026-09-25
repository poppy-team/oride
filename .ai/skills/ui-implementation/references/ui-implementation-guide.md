# Design Tokens & Component State Matrix Engineering Guide

## 1. Design Tokens Community Group (DTCG 2025.10) Standard
Design tokens are the visual atoms of the design system. Prumo adopts the DTCG 2025.10 specification for design tokens:
- **Primitive Tokens**: Direct raw values (`color.blue.500 = "#0066cc"`). Never consumed directly by UI components.
- **Semantic Tokens**: Intent-based aliases (`color.surface.brand = "{color.blue.500}"`). Components consume these.
- **Component Tokens**: Scoped overrides (`button.primary.background = "{color.surface.brand}"`).

## 2. The 19-Point State Matrix
A component is not complete when it only renders the "happy path" default mock. The table below defines the full state matrix required for complex UI:

| State | Trigger | Visual Cue | A11y Requirement |
|---|---|---|---|
| `default` | Initial render | Neutral baseline | Normal tab order |
| `hover` | Pointer enters element | Surface luminance shift | `cursor: pointer` |
| `focus-visible` | Keyboard tab navigation | 2px solid ring, 2px offset | High contrast (>= 3:1) |
| `pressed` | Mouse down / Key down | Scale or darker surface | Tactile feedback |
| `disabled` | Business rule | 50% opacity, greyed out | `aria-disabled="true"` |
| `loading` | Async in flight | Spinner / Skeleton pulse | `aria-busy="true"` |
| `empty` | Zero items in dataset | Clean icon + call-to-action | Helpful text |
| `error` | Validation rejected | Red border + error message | `aria-invalid="true"` |
| `long-content` | German/Spanish locale | Ellipsis + Tooltip | Container overflow hidden |

## 3. Keyboard & Focus Management Patterns
- **Buttons / Links**: Space / Enter to activate.
- **Menus / Dropdowns**: Escape to close and return focus to toggle button.
- **Composite Widgets (Tabs, Radios)**: Roving tabindex: Only active tab has `tabindex="0"`, inactive tabs have `tabindex="-1"`. Arrow keys cycle through tabs.
