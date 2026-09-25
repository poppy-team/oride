# UI Component Implementation & Verification Report

## 1. Component Metadata
- **Component Name**: `<ComponentName>`
- **Framework / Runtime**: React / Svelte / Vue / Rust (egui) / Go (TUI)
- **Design Tokens Applied**: DTCG 2025.10 Semantic Tokens
- **Date**: YYYY-MM-DDTHH:MM:SSZ
- **Author / Agent**: `<agent-id>`

## 2. Anatomy & Props Specification
- **Variants Supported**: `primary`, `secondary`, `danger`, `ghost`
- **Sizes Supported**: `sm`, `md`, `lg`
- **Slots / Children**: Icon slot, Label slot, Flyout overlay

## 3. 19-Point State Matrix Audit Results
| State | Verified | Visual Fixture Available | Notes |
|---|---|---|---|
| `default` | YES / NO | YES / NO | |
| `hover` | YES / NO | YES / NO | |
| `focus-visible` | YES / NO | YES / NO | 2px ring verified |
| `pressed / active` | YES / NO | YES / NO | |
| `disabled` | YES / NO | YES / NO | Pointer events suppressed |
| `loading` | YES / NO | YES / NO | aria-busy set |
| `empty` | YES / NO | YES / NO | |
| `error` | YES / NO | YES / NO | Descriptive error text |
| `locale-expansion` | YES / NO | YES / NO | Tested with 200% string length |

## 4. Accessibility & Keyboard Test Results
- **Semantic Element**: `<button>` / `<input>` / AccessKit `WidgetInfo`
- **Keyboard Navigation**: Tab, Enter, Space, Escape verified
- **Contrast Ratio (Focus Ring)**: X.X:1 (>= 3:1 required)
- **Touch Target Size**: W x H px (>= 24x24px required)

## 5. Verification Sign-Off
- [ ] Component passes all state matrix tests and is token-compliant.
