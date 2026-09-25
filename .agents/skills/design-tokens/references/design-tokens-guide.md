# W3C Design Tokens & Theming — Technical Reference Guide

## 1. W3C DTCG Format Specification (2025.10)
Design tokens are the visual design atoms of the design system. The W3C Design Tokens Community Group (DTCG) specifies a vendor-neutral JSON format:

```json
{
  "color": {
    "brand": {
      "primary": {
        "$value": "#2563eb",
        "$type": "color",
        "$description": "Core brand interactive color"
      }
    }
  }
}
```

### Supported DTCG `$type` Primitives:
- `color`: Hex (`#RRGGBB` / `#RRGGBBAA`), RGB, or modern perceptual Oklch strings.
- `dimension`: Values with explicit unit suffix (`rem`, `px`, `em`, `%`).
- `duration`: Time string with unit (`ms`, `s`).
- `cubicBezier`: Array of 4 numbers `[x1, y1, x2, y2]` representing timing function.
- `number`: Unitless scalar values (e.g. line-height multipliers, z-indices).
- `fontFamily`: String or array of strings for font fallbacks.
- `fontWeight`: Numerical font weight (`100` to `900`) or standard keyword.
- `shadow`: Object with `color`, `offsetX`, `offsetY`, `blur`, and `spread`.
- `typography`: Composite token combining `fontFamily`, `fontSize`, `lineHeight`, `fontWeight`.

---

## 2. Three-Tier Token Topology

```
+-------------------------------------------------------------------------+
| Tier 1: Primitive Tokens (Global)                                       |
| - Pure values, zero contextual intent                                  |
| - e.g.: color.blue.500, spacing.16, radius.full, font.sans             |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Tier 2: Semantic Tokens (System / Mode)                                 |
| - Express design intent and purpose; values are aliases to Tier 1       |
| - e.g.: color.action.primary.default = {color.blue.500}                |
| - Swapped seamlessly between Light and Dark themes                      |
+-------------------------------------------------------------------------+
                                    |
                                    v
+-------------------------------------------------------------------------+
| Tier 3: Component Tokens (Scoped)                                       |
| - Scoped to a specific component boundary                               |
| - e.g.: button.primary.background = {color.action.primary.default}     |
| - Allows targeted overrides without breaking system-wide semantics     |
+-------------------------------------------------------------------------+
```

---

## 3. WCAG 2.2 Contrast Ratio Mathematics
WCAG 2.2 defines the contrast ratio between two colors based on relative luminance $L$:
$$L = 0.2126 \cdot R' + 0.7152 \cdot G' + 0.0722 \cdot B'$$
Where $C'$ is the linearized color component:
$$C' = \begin{cases} \frac{C}{12.92} & \text{if } C \le 0.04045 \\ \left(\frac{C + 0.055}{1.055}\right)^{2.4} & \text{if } C > 0.04045 \end{cases}$$
The contrast ratio between lighter color $L_1$ and darker color $L_2$ is:
$$\text{Contrast Ratio} = \frac{L_1 + 0.05}{L_2 + 0.05}$$

### WCAG 2.2 Thresholds:
- **Level AA Normal Text**: $\ge 4.5:1$
- **Level AA Large Text ($\ge 18\text{pt}$ or $\ge 14\text{pt}$ bold)**: $\ge 3.0:1$
- **User Interface Components & Graphical Objects**: $\ge 3.0:1$
- **Level AAA Normal Text**: $\ge 7.0:1$

---

## 4. Multi-Platform Compilation Pipeline
Token compilers resolve aliases and serialize targets:
1. **Web (CSS Custom Properties)**:
   ```css
   :root {
     --color-action-primary-default: #2563eb;
     --spacing-4: 1rem;
   }
   ```
2. **Native Rust / Desktop Engine**:
   ```rust
   pub mod tokens {
       pub const COLOR_ACTION_PRIMARY_DEFAULT: egui::Color32 = egui::Color32::from_rgb(37, 99, 235);
   }
   ```
