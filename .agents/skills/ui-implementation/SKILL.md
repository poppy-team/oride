---
name: ui-implementation
description: Component-driven UI implementation, multi-state matrix (default, hover, focus-visible, active, disabled, loading, empty, error), DTCG design tokens, responsive layout contracts, and accessibility tree mapping.
---

# Design System UI Component Implementation

## 1. Title and Description
**Design System UI Component Implementation (`ui-implementation`)**
Governs the creation, refactoring, and verification of production UI components across Web (React, Svelte, Vue, native HTML/CSS), Desktop Native (egui, Slint), and Terminal TUI (Bubble Tea v2, Lip Gloss). Enforces full component state matrices, Design Tokens (DTCG 2025.10), responsive layout constraints, and deterministic rendering.

## 2. Purpose
Eliminate broken UI edge states, unstyled loading/error flashes, brittle layout overflows, inaccessible interactive elements, and hardcoded styling by transforming design intent into rigorously tested, token-bound components.

## 3. Prerequisites
- Component wireframe or design specification.
- Design tokens repository or token variables (CSS custom properties, Rust/Go theme structs).
- Target UI framework / runtime toolchain.

## 4. Inputs
- Component specification (props, slots, variants, sizes, events).
- Design token set (colors, typography, spacing, border radii, elevation).
- Accessibility requirements (ARIA role, keyboard map, AccessKit widget info).

## 5. Outputs
- Fully implemented, modular UI component source code.
- Interactive component fixture or story (Storybook, Ladle, `egui_kittest`, or TUI preview).
- Complete state matrix test suite verifying all 19 component states.
- Implementation report conforming to `templates/ui-component-report.md`.

## 6. Execution Steps

### Step 1: Formalize Component Anatomy
Decompose the component into a clean, hierarchical structure:
1. **Root / Container**: Bounds the component, manages outer layout flow, and handles tab stop / focus ring.
2. **Leading / Trailing Adornments**: Icons, avatars, status badges, or action buttons.
3. **Content Area**: Primary label, body text, or slot for arbitrary child elements.
4. **Overlay / Flyout** (if applicable): Popovers, dropdown menus, tooltips anchored to the trigger.

### Step 2: Bind to Design Tokens (DTCG 2025.10)
Never hardcode raw hex values, pixel sizes, or arbitrary font names. Bind strictly to semantic design tokens:
- **Surfaces**: `color.surface.default`, `color.surface.hover`, `color.surface.active`.
- **Text / Foreground**: `color.text.primary`, `color.text.secondary`, `color.text.disabled`.
- **Borders & Focus**: `color.border.subtle`, `color.focus.ring` (min 2px thickness with contrast).
- **Spacing Rhythm**: Multiples of the 4px / 8px scale (`space.1` = 4px, `space.2` = 8px, `space.4` = 16px).
- **Typography**: `font.family.base`, `font.size.sm`, `font.lineHeight.normal`.

### Step 3: Implement the 19-Point State Matrix
Every interactive component must account for the following states:
1. `default`: Rest state ready for user interaction.
2. `hover`: Visual affordance indicating pointer presence.
3. `focus-visible`: High-contrast keyboard focus indicator (never suppress without replacement!).
4. `pressed / active`: Immediate tactile feedback during mouse down or key press.
5. `selected / checked`: Persistent toggle or selection state.
6. `expanded / open`: Menu or disclosure state with appropriate ARIA attributes.
7. `disabled`: Visually muted, pointer events blocked, removed from tab order or `aria-disabled="true"`.
8. `read-only`: Non-editable but focusable and readable by screen readers.
9. `loading / pending`: Integrated spinner or skeleton placeholder with `aria-busy="true"`.
10. `empty`: Structured visual cue when zero data items are available.
11. `error`: Validation failure with descriptive error message and error border token.
12. `warning`: Cautionary feedback that does not block form submission.
13. `success`: Confirmation feedback on completed actions.
14. `offline`: Graceful notification when network connectivity is lost.
15. `permission-denied`: Informative placeholder when user lacks sufficient rights.
16. `drag / drop`: Visual highlight during drag-and-drop operations.
17. `destructive-confirmation`: Two-step verification state before irrecoverable actions.
18. `long-content / locale expansion`: Text truncation with tooltip or wrapping without breaking parent boundaries.
19. `high-contrast mode`: Compatibility with Windows High Contrast Mode or forced-colors CSS.

### Step 4: Responsive & Adaptive Layout Contract
1. **Container Queries Over Breakpoints**: Prefer container queries (`@container`) for modular components so they adapt to their parent column width rather than global screen width.
2. **Defensive Sizing**: Define `min-width`, `max-width`, and `overflow: hidden` / `text-overflow: ellipsis` on labels to prevent container blowout when text expands in German/Spanish.
3. **Touch Targets**: Minimum interactive touch target must be **24x24px** (WCAG 2.2 Target Size Minimum) and preferably **44x44px** on mobile platforms.

### Step 5: Accessibility & Keyboard Bindings
1. **Native Semantics**: Use native `<button>`, `<a>`, `<input>` elements before custom divs.
2. **Custom Renderers (egui / Slint)**: Register `WidgetInfo` with role, name, and current value to feed AccessKit/AT-SPI.
3. **Keyboard Handlers**:
   - `Enter` / `Space`: Activate button or toggle checkbox.
   - `Escape`: Close active popover, menu, or modal and restore focus to trigger.
   - Arrow keys: Navigate through composite items with roving tabindex.

### Step 6: Testing & Visual Fixture Generation
1. Implement a component showcase story covering all primary variants and states.
2. Author automated component tests asserting state transitions and keyboard interactions.

## 7. Verification
```bash
# 1. Component unit test pass
npm test -- src/components/Button.test.tsx
# Or for Rust/egui:
cargo test -p ui_components

# 2. State matrix audit script
./scripts/audit_states.sh src/components/Button.tsx
```

## 8. Fallbacks & Failure Recovery
| Failure Class | Root Cause | Deterministic Remediation |
|---|---|---|
| Text Truncation / Layout Blowout | String longer than container breaks layout | Apply `truncate` (flex-shrink, overflow: hidden, text-overflow: ellipsis) and render full string in accessible tooltip. |
| Missing Focus Ring | Global `outline: none` removes indicator | Apply `:focus-visible { outline: 2px solid var(--color-focus-ring); outline-offset: 2px; }`. |
| Layout Shift on Loading (CLS) | Component expands abruptly when async data arrives | Pre-render skeleton placeholder matching exact dimensions of final loaded content. |
| ARIA State Desynchronization | Visual toggle changes without updating `aria-expanded` | Bind `aria-expanded={isOpen}` directly to the reactive boolean state variable. |

## 9. Constraints
- **Zero raw colors**: Hardcoded hex, rgb, or hsl values are forbidden; all values must resolve through design tokens.
- **Zero mouse-only interactions**: Every feature accessible by mouse must have 100% functional parity via keyboard.
- **No missing states**: A component PR without explicit definitions for loading, empty, and error states will fail review.

## 10. Examples

### Anti-Pattern: Div Button with Hardcoded Styles & Missing States
```tsx
// BAD: Inaccessible div-button, hardcoded colors, no keyboard support, missing states
export const BadButton = ({ onClick, label }: { onClick: () => void; label: string }) => {
    return (
        <div 
            onClick={onClick}
            style={{ backgroundColor: '#0070f3', color: '#fff', padding: '10px 20px', borderRadius: '4px' }}
        >
            {label}
        </div>
    );
};
```

### Idiomatic Pattern: Token-Bound Accessible Component with Multi-State
```tsx
// GOOD: Native button, design tokens, focus-visible ring, loading state, keyboard support
import React from 'react';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'primary' | 'secondary' | 'danger';
    size?: 'sm' | 'md' | 'lg';
    isLoading?: boolean;
    icon?: React.ReactNode;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(({
    variant = 'primary',
    size = 'md',
    isLoading = false,
    disabled = false,
    icon,
    children,
    className = '',
    ...rest
}, ref) => {
    const isInteractive = !disabled && !isLoading;

    return (
        <button
            ref={ref}
            disabled={disabled || isLoading}
            aria-busy={isLoading}
            className={`
                inline-flex items-center justify-center font-medium transition-colors
                rounded-[var(--radius-md)] border border-transparent
                focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-focus-ring)] focus-visible:ring-offset-2
                disabled:cursor-not-allowed disabled:opacity-50
                ${variant === 'primary' ? 'bg-[var(--color-surface-brand)] text-[var(--color-text-on-brand)] hover:bg-[var(--color-surface-brand-hover)]' : ''}
                ${variant === 'secondary' ? 'bg-[var(--color-surface-secondary)] text-[var(--color-text-primary)] hover:bg-[var(--color-surface-secondary-hover)] border-[var(--color-border-subtle)]' : ''}
                ${variant === 'danger' ? 'bg-[var(--color-surface-danger)] text-[var(--color-text-on-danger)] hover:bg-[var(--color-surface-danger-hover)]' : ''}
                ${size === 'sm' ? 'px-2.5 py-1.5 text-xs min-h-[32px]' : ''}
                ${size === 'md' ? 'px-4 py-2 text-sm min-h-[40px]' : ''}
                ${size === 'lg' ? 'px-6 py-3 text-base min-h-[48px]' : ''}
                ${className}
            `}
            {...rest}
        >
            {isLoading ? (
                <span className="mr-2 inline-block h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" aria-hidden="true" />
            ) : icon ? (
                <span className="mr-2 inline-flex" aria-hidden="true">{icon}</span>
            ) : null}
            <span>{children}</span>
        </button>
    );
});
Button.displayName = 'Button';
```
