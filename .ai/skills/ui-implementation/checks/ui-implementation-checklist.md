# UI Component Implementation Verification Checklist

## 1. Design Token Integration
- [ ] Zero raw color literals (#hex, rgb) used in styles; all colors resolve to semantic design tokens.
- [ ] Spacing values adhere to the design token scale (multiples of 4px/8px).
- [ ] Border radius, font families, and shadows map to design tokens.

## 2. 19-Point State Matrix Audit
- [ ] Default / Normal state rendered and verified.
- [ ] Hover state provides clear visual affordance.
- [ ] Focus-visible indicator is present with minimum 2px thickness and >= 3:1 contrast ratio.
- [ ] Active / Pressed state provides immediate tactile feedback.
- [ ] Disabled state sets `disabled` / `aria-disabled="true"` and suppresses pointer events.
- [ ] Loading / Pending state displays spinner or skeleton with `aria-busy="true"`.
- [ ] Empty state provides helpful guidance when data is empty.
- [ ] Error state displays descriptive error message and error border.
- [ ] Long content and internationalization expansion handles text without layout explosion.

## 3. Accessibility & Keyboard Operability
- [ ] Operable entirely via keyboard (`Enter`, `Space`, `Escape`, arrow keys where applicable).
- [ ] Semantic HTML element used (`<button>`, `<a>`, `<input>`) or full ARIA role provided.
- [ ] ARIA states (`aria-expanded`, `aria-checked`, `aria-selected`) synchronized with reactive state.
- [ ] Minimum touch target size meets WCAG 2.2 criteria (>= 24x24px, preferably >= 44x44px).

## 4. Testing & Code Quality
- [ ] Component unit tests cover all primary variants and state transitions.
- [ ] Component preview fixture (Storybook/Ladle/kittest) exists.
- [ ] Zero linting or TypeScript compilation errors.
