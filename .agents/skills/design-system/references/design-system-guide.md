# Enterprise Design System Architecture & Engineering Guide

## 1. Design Token Architecture & Pipeline

Design tokens represent the single source of truth for all visual decisions across platforms (Web, iOS, Android, Native Desktop).

### 1.1 DTCG Specification Structure

The Design Tokens Community Group (W3C DTCG) defines the standard token interchange format:

```json
{
  "color": {
    "palette": {
      "blue": {
        "500": { "$value": "#2563eb", "$type": "color" },
        "600": { "$value": "#1d4ed8", "$type": "color" }
      }
    },
    "action": {
      "primary": {
        "background": {
          "$value": "{color.palette.blue.500}",
          "$type": "color",
          "$description": "Primary interactive element surface background"
        },
        "backgroundHover": {
          "$value": "{color.palette.blue.600}",
          "$type": "color"
        }
      }
    }
  }
}
```

### 1.2 Compilation Targets

Using toolchains such as Style Dictionary, tokens compile into:
- **Web**: CSS Custom Properties (`:root { --color-action-primary-background: #2563eb; }`), SCSS variables, TypeScript typed constants.
- **iOS**: Swift structs or Asset Catalogs (`Color("actionPrimaryBackground")`).
- **Android**: XML resources (`<color name="action_primary_background">#2563eb</color>`) or Jetpack Compose Color tokens.

---

## 2. Component Architecture: Headless & Polymorphic

### 2.1 The Polymorphic Slot (`asChild`) Pattern

To avoid creating tag soups or non-semantic HTML (such as nesting `<button>` inside `<a>` or rendering clickable `<div>` elements), components should implement the Slot pattern (popularized by Radix UI):

```tsx
import React, { cloneElement, isValidElement } from 'react';

export interface SlotProps extends React.HTMLAttributes<HTMLElement> {
  children?: React.ReactNode;
}

export const Slot = React.forwardRef<HTMLElement, SlotProps>((props, ref) => {
  const { children, ...slotProps } = props;
  if (isValidElement(children)) {
    return cloneElement(children, {
      ...slotProps,
      ...children.props,
      // Merge refs and event handlers properly
      ref: ref ? ref : (children as any).ref,
      className: [slotProps.className, children.props.className].filter(Boolean).join(' '),
    });
  }
  return null;
});
```

### 2.2 Variant Contract with Type-Safe Enums

Components should declare clear visual variant contracts:

```typescript
export type ButtonVariant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive';
export type ButtonSize = 'sm' | 'md' | 'lg';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  asChild?: boolean;
  isLoading?: boolean;
}
```

---

## 3. Accessibility Guarantees

1. **Focus Ring Contrast**: Interactive focus rings must satisfy at least 3:1 contrast against both the component background and the adjacent page background.
2. **Keyboard Handlers**: Components handling custom keyboard events must never swallow global modifier keys (`Ctrl`, `Meta`, `Alt`, `Shift`) unless explicitly handling shortcuts.
3. **No Trapping Without Intention**: Modals and slide-out sheets must trap focus while open and release/restore focus to the triggering element upon closing. Non-modal dialogs must never trap focus.

---

## 4. Testing & Regression Strategy

```
          / \
         /   \      Visual Regression Tests (Playwright / Chromatic)
        / Visual\   - Multi-theme screenshot diffs across all story states
       /---------\
      / A11y Tests\ Automated Accessibility Scans (Axe-core, Pa11y)
     /-------------\ - Zero WCAG AA violations on every story
    /   Unit Tests  \ Component Logic & Contract Tests (Jest / Vitest)
   /-----------------\ - Keyboard interaction, event handlers, prop validation
```

---

## 5. Versioning and Deprecation Strategy

1. **Semantic Versioning**:
   - **Patch**: Styling bug fixes, token color adjustments without contract changes.
   - **Minor**: New components, non-breaking additional props or variant additions.
   - **Major**: Renaming or removing components/props, changing default variant behavior.
2. **Deprecation Grace Period**:
   - Deprecated props must remain active for at least one major version cycle.
   - Emit a developer console warning: `console.warn('[DesignSystem] Prop "isPrimary" is deprecated. Use variant="primary" instead.')`.
   - Provide AST-based codemods (using `jscodeshift` or `ts-morph`) for automated customer upgrades.
