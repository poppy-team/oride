# Component Specification: {{Component.Name}}

## 1. Overview & Semantic Intent
- **Component ID**: `{{component-id}}`
- **Category**: [Primitive | Composite | Layout | Feedback]
- **Target Platforms**: [Web | iOS | Android | Cross-Platform]
- **Description**: {{Brief description of the component purpose and user interaction model}}

---

## 2. API Contract & Props

```typescript
export interface {{Component.Name}}Props {
  /** Visual variant styling */
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive';
  /** Dimension sizing */
  size?: 'sm' | 'md' | 'lg';
  /** Polymorphic slot rendering delegate */
  asChild?: boolean;
  /** Disabled state */
  disabled?: boolean;
  /** Busy/loading indicator state */
  isLoading?: boolean;
  /** Component content */
  children?: React.ReactNode;
}
```

---

## 3. Design Token Mappings

| Element / Layer | Token Name | Default Value | Fallback / Purpose |
|-----------------|------------|---------------|--------------------|
| Container Background | `--color-action-{{variant}}-bg` | `var(--color-blue-600)` | Background surface |
| Text Color | `--color-action-{{variant}}-fg` | `var(--color-white)` | Foreground text contrast |
| Border | `--color-action-{{variant}}-border`| `transparent` | Border boundary |
| Focus Ring | `--color-focus-ring` | `var(--color-blue-500)` | Focus outline |
| Radius | `--radius-md` | `0.375rem (6px)` | Corner rounding |
| Spacing Padding | `--space-btn-{{size}}` | `0.5rem 1rem` | Internal padding |

---

## 4. Interaction & State Matrix

| State | Visual Treatment | ARIA Attributes | Keyboard Behavior |
|-------|------------------|-----------------|-------------------|
| **Default** | Standard token values | None | Focusable via `Tab` |
| **Hover** | `--color-action-*-bg-hover` | None | Cursor: pointer |
| **Focus-Visible** | 2px solid `--color-focus-ring`, 2px offset | None | Shows visible ring |
| **Active/Pressed** | `--color-action-*-bg-active` | None | Depressed effect |
| **Disabled** | 40% opacity, no hover effect | `aria-disabled="true"` | Non-clickable |
| **Loading** | Spinner icon replaces/prepends icon | `aria-busy="true"` | Clicks suppressed |

---

## 5. Accessibility & ARIA Contract
- **Role**: `button` (or inherited from slotted element when `asChild=true`).
- **Keyboard Handling**:
  - `Enter`: Activates action.
  - `Space`: Activates action on keyup.
- **Contrast Ratios**: Verified $\ge 4.5:1$ against surface background across all supported color themes.
- **Target Size**: Minimum $44 \times 44\text{ px}$ clickable area (or minimum $24 \times 24\text{ px}$ with spacing per WCAG 2.2 SC 2.5.8).

---

## 6. Verification & Test Plan
- [ ] TypeScript prop contract validation.
- [ ] Automated Axe accessibility scan (0 violations).
- [ ] Storybook story created covering all variants and states.
- [ ] Playwright visual regression baseline captured in light and dark mode.
