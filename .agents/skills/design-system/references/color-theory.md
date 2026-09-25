# Color Theory in Design Systems

A robust design system relies heavily on a structured color palette. Understanding color theory ensures accessible, aesthetically pleasing, and functional interfaces.

## 1. Color Models
- **RGB / HEX:** Used for digital screens (Red, Green, Blue).
- **HSL / HSLA:** Hue, Saturation, Lightness. Excellent for programmatically generating palettes by keeping Hue constant and varying Lightness.
- **OKLCH:** A perceptually uniform color space. Ideal for design systems to ensure consistent contrast ratios across different hues.

## 2. Palette Roles
- **Primary:** The brand color. Used for primary actions (e.g., CTA buttons).
- **Secondary / Accent:** Complements the primary color. Used to highlight secondary information.
- **Neutral / Grayscale:** Used for typography, borders, backgrounds. Usually 8-10 shades from almost white to almost black.
- **Semantic Colors:**
  - **Success:** Green (Confirmations, positive trends)
  - **Warning:** Yellow/Orange (Alerts requiring attention)
  - **Danger/Error:** Red (Destructive actions, validation errors)
  - **Info:** Blue (Neutral system feedback)

## 3. Accessibility & Contrast
According to WCAG guidelines, text must have a sufficient contrast ratio against its background:
- **AA Level:** Minimum 4.5:1 for normal text, 3.1:1 for large text.
- **AAA Level:** Minimum 7.1:1 for normal text, 4.5:1 for large text.

Always test text/background combinations during the creation of your color tokens.
