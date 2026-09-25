# W3C Design Tokens Architecture & Multi-Platform Compilation

## 1. Purpose
Design, maintain, validate, and compile cross-platform design tokens adhering strictly to the **W3C Design Tokens Community Group (DTCG 2025.10)** specification. This skill enforces a 3-tier token topology (Primitive $\rightarrow$ Semantic $\rightarrow$ Component), automated WCAG 2.2 contrast compliance across light and dark themes, eradication of hardcoded aesthetic values in component implementations, and automated code generation for CSS Custom Properties, TypeScript constants, and native Rust/C++ configurations.

---

## 2. Use When
- Establishing or modifying design token definitions (color palettes, typography scales, spacing grids, corner radii, elevation shadows, motion curves).
- Designing multi-theme systems (light, dark, high-contrast, brand variants) using semantic aliasing.
- Building or auditing automated token compilation pipelines (Style Dictionary, DTCG compilers).
- Validating contrast ratios between semantic foreground and background token pairs.
- Operating in mode(s): `implementation`, `design`, `review`, `audit`.

---

## 3. Do Not Use When
- Writing low-level GPU fragment shaders or hardware vertex layout descriptors (use `shaders`).
- Implementing raw immediate-mode canvas primitives without design system contracts.
- Building disposable throwaway mockups that explicitly do not require design tokens.

---

## 4. Required Context
Before declaring or compiling design tokens, verify:
- **DTCG Schema Conformance**: Format requires `$value`, `$type`, and optional `$description` attributes.
- **Three-Tier Hierarchy Mapping**:
  1. *Tier 1 (Primitive/Global)*: Raw values without context (`color.slate.900: { "$value": "#0f172a", "$type": "color" }`).
  2. *Tier 2 (Semantic/Alias)*: Purpose-driven abstractions (`color.surface.canvas: { "$value": "{color.slate.900}", "$type": "color" }`).
  3. *Tier 3 (Component)*: Scoped variables (`button.primary.background: { "$value": "{color.surface.canvas}", "$type": "color" }`).
- **Target Export Platforms**: CSS custom properties, TypeScript const objects, Rust enums/structs, or Android XML / iOS Swift.

---

## 5. Procedure

```
[Design Intent / Figma Tokens]
         |
         v
[1. DTCG JSON Schema Authoring] -----> Declare Primitives with explicit $type
         |
         v
[2. Semantic & Mode Aliasing] -------> Map Light/Dark semantic aliases {token.ref}
         |
         v
[3. Contrast & Math Validation] -----> Audit 4.5:1 text and 3:1 focus ratios
         |
         v
[4. Multi-Target Compilation] -------> Generate CSS, TS, Rust, C++ headers
         |
         v
[5. Component Invariant Check] ------> Verify 0 raw hex colors in UI source
```

### Step 1: Declare Primitive Tokens with Explicit Types
Every token definition must specify its formal DTCG `$type`:
`color`, `dimension`, `duration`, `cubicBezier`, `number`, `fontFamily`, `fontWeight`, `shadow`, `typography`.
```json
{
  "color": {
    "neutral": {
      "0": { "$value": "#ffffff", "$type": "color" },
      "900": { "$value": "#0f172a", "$type": "color" }
    }
  },
  "spacing": {
    "sm": { "$value": "0.5rem", "$type": "dimension" },
    "md": { "$value": "1rem", "$type": "dimension" }
  }
}
```

### Step 2: Semantic Mapping & Theme Invariance
Components must never bind directly to primitive tokens. Bind components exclusively to semantic aliases:
```json
{
  "semantic": {
    "color": {
      "surface": {
        "canvas": {
          "$value": "{color.neutral.0}",
          "$type": "color",
          "$description": "Main background surface"
        }
      },
      "text": {
        "primary": {
          "$value": "{color.neutral.900}",
          "$type": "color"
        }
      }
    }
  }
}
```
In dark mode, override `{semantic.color.surface.canvas}` to `{color.neutral.900}` and `{semantic.color.text.primary}` to `{color.neutral.0}`.

### Step 3: Automated Contrast Verification
Verify that every semantic foreground token paired with a background token satisfies WCAG 2.2 contrast invariants:
- Body text: $\ge 4.5:1$.
- Large text ($\ge 18\text{pt}$ or $\ge 14\text{pt}$ bold): $\ge 3:1$.
- Interactive focus rings and icons: $\ge 3:1$.

### Step 4: Multi-Target Code Generation
Compile tokens deterministically:
1. **CSS Output**:
   ```css
   :root {
     --color-surface-canvas: #ffffff;
     --color-text-primary: #0f172a;
     --spacing-md: 1rem;
   }
   [data-theme="dark"] {
     --color-surface-canvas: #0f172a;
     --color-text-primary: #ffffff;
   }
   ```
2. **TypeScript / Rust Output**: Emit immutable, strongly typed structures for native codebases.

### Step 5: Static Component Linting
Run static scanners across all frontend / UI components to guarantee zero raw color literals (`#fff`, `rgb(...)`) exist outside design token definitions.

---

## 6. Decision Rules & Invariants
- **RULE 1 (No Component-to-Primitive Binding)**: UI components must reference Semantic Tokens (`var(--color-surface-canvas)`), never raw Primitive Tokens (`var(--color-neutral-0)`).
- **RULE 2 (Contrast Conformance)**: Any semantic text/background pairing with contrast $< 4.5:1$ is rejected at build time.
- **RULE 3 (Strict DTCG Syntax)**: Token JSON files must strictly adhere to the `$value` and `$type` attributes of W3C DTCG specification.
- **RULE 4 (Relative Units for Dimensions)**: Spacing and typography tokens must use relative units (`rem`, `ch`, `em`) rather than hardcoded physical pixels (`px`) to support user zoom and reflow.

---

## 7. Evidence Required
- **Schema Validation Evidence**: Clean validation of token JSON against the DTCG schema.
- **Contrast Audit Matrix**: Mathematical evidence proving all semantic foreground/background pairs meet WCAG AA requirements.
- **Hardcoded Aesthetic Audit**: Zero raw hex or pixel strings detected in production UI components.

---

## 8. Output Contract
- Canonical DTCG Token definitions (`tokens/*.json`).
- Compiled CSS Custom Properties (`dist/tokens.css`) and TypeScript bindings (`dist/tokens.ts`).
- Design Token Audit & Contrast Report (`templates/design-tokens-report.md`).

---

## 9. Stop Conditions
- All token files validate against DTCG schema.
- All theme contrast checks pass with $\ge 4.5:1$ score.
- Token budget exhausted.
- Blocked on external dependency.

---

## 10. Escalation Rules
- Escalate to brand lead if brand guideline primary colors fail WCAG AA contrast against required background surfaces.
- Escalate if multi-brand theming requirements exceed token compiler alias resolution capabilities.
