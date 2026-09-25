# WCAG 2.2 Color Contrast & Visual Legibility Checklist

## 1. Text Contrast (WCAG 2.2 SC 1.4.3 & 1.4.6)
- [ ] Regular body text ($< 18\text{pt}$ or $< 14\text{pt}$ bold) achieves $\ge 4.5:1$ contrast against adjacent background.
- [ ] Large text ($\ge 18\text{pt}$ / $\ge 24\text{px}$ or $\ge 14\text{pt}$ bold / $\ge 18.5\text{px}$ bold) achieves $\ge 3.0:1$ contrast.
- [ ] Incidental text, decorative text, or text that is part of an inactive UI component is properly classified as exempt.
- [ ] Text rendered over images or gradients uses a solid backing scrim or overlay guaranteeing minimum contrast at all points.

## 2. Non-Text Contrast (WCAG 2.2 SC 1.4.11)
- [ ] Visual boundaries of form inputs, buttons, and selectable cards achieve $\ge 3.0:1$ contrast against surrounding surfaces.
- [ ] Graphical icons and symbols essential for understanding achieve $\ge 3.0:1$ contrast against adjacent backgrounds.
- [ ] State indicators (selected, expanded, active tabs) achieve $\ge 3.0:1$ contrast against unselected states.

## 3. Focus Indicator Legibility (WCAG 2.2 SC 2.4.11 & 2.4.13)
- [ ] Focus indicator has an area of at least 2 CSS pixels around the perimeter.
- [ ] Focus indicator achieves $\ge 3.0:1$ contrast against the unfocused state.
- [ ] Focus indicator achieves $\ge 3.0:1$ contrast against the focused component background.

## 4. Use of Color & Multi-Hue Independence (SC 1.4.1)
- [ ] Color is never the sole visual cue to indicate errors, successes, warnings, or interactive links.
- [ ] Inline links inside paragraphs have an underline or a $\ge 3.0:1$ contrast ratio against surrounding body text.
- [ ] Form error states display text error messages and error icons in addition to red border styling.
- [ ] Interface maintains full comprehensibility under Protanopia, Deuteranopia, and Tritanopia simulations.

## 5. Alpha Compositing & Dynamic State Hygiene
- [ ] Translucent colors ($rgba$, $hsla$) are composited mathematically against underlying surfaces before checking contrast.
- [ ] Contrast ratios re-verified across all hover (`:hover`), active (`:active`), and focus (`:focus-visible`) states.
- [ ] Light and dark themes pass contrast checks independently without relying on browser auto-inversion.
