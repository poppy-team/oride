# Color Contrast Verification & Perceptual Legibility (WCAG 2.2 & APCA)

## 1. Purpose
Verify, audit, and mathematically validate color contrast ratios across all user interfaces, graphical components, and interactive states. This skill enforces strict compliance with **WCAG 2.2 Success Criteria 1.4.3 (Contrast Minimum), 1.4.11 (Non-text Contrast), and 1.4.1 (Use of Color)**, integrates the Advanced Perceptual Contrast Algorithm (APCA), validates alpha-blended composite surfaces, and verifies multi-hue legibility under color blindness simulations (Protanopia, Deuteranopia, Tritanopia).

---

## 2. Use When
- Auditing UI designs, design tokens, web pages, or desktop application themes for accessibility compliance.
- Verifying focus ring indicators, input borders, button states, and graphical charts against adjacent surfaces.
- Calculating effective contrast on semi-transparent overlays, glassmorphic surfaces, or gradient backgrounds.
- Testing that color is never the sole visual indicator of status, errors, or critical alerts.
- Operating in mode(s): `audit`, `testing`, `review`, `implementation`.

---

## 3. Do Not Use When
- Designing structural keyboard focus trapping or roving tabindex order (use `keyboard-accessibility` or `focus-management`).
- Auditing screen reader tree semantics and aria-live announcements (use `screen-reader`).
- Developing GPU shaders without UI or text rendering.

---

## 4. Required Context
Before verifying contrast ratios, verify:
- **Target Standard**: WCAG 2.2 Level AA (Mandatory: $4.5:1$ text, $3:1$ UI components/large text) or Level AAA ($7:1$ text, $4.5:1$ large text).
- **Text Classification**: Normal body text ($< 18\text{pt}$ or $< 14\text{pt}$ bold) vs Large text ($\ge 18\text{pt}$ / $\ge 24\text{px}$ or $\ge 14\text{pt}$ bold / $\ge 18.5\text{px}$ bold).
- **Interactive State Matrix**: Default, Hover, Focus-Visible, Active, Disabled, Error, and Warning states.
- **Surface Stack**: Composition stack for alpha blending ($C_{\text{foreground}}$ on $C_{\text{overlay}}$ on $C_{\text{background}}$).

---

## 5. Procedure

```
[UI Design / Token Palette]
         |
         v
[1. Surface Alpha Compositing] -----> Flatten translucent layers to opaque sRGB
         |
         v
[2. Relative Luminance Calculation] -> Linearize RGB components and compute L1, L2
         |
         v
[3. WCAG 2.2 AA/AAA Ratio Check] ---> Verify 4.5:1 (text) and 3.0:1 (UI / focus)
         |
         v
[4. APCA Perceptual Legibility] ----> Check Lc score against font weight & size
         |
         v
[5. Color Blindness & Dual Cue Audit] Protanopia/Deuteranopia simulation + non-color cue
```

### Step 1: Alpha Flattening & Surface Composition
When text or backgrounds have alpha transparency ($A < 1.0$), compute the effective composite color $C_{\text{effective}}$ before calculating luminance:
$$C_{\text{effective}} = C_{\text{fg}} \times A_{\text{fg}} + C_{\text{bg}} \times (1 - A_{\text{fg}})$$
If multiple translucent layers exist, composite them sequentially from background to foreground.

### Step 2: Calculate Relative Luminance & WCAG Ratio
1. Linearize standard sRGB channels ($R, G, B \in [0, 1]$):
   $$C_{\text{linear}} = \begin{cases} \frac{C_{\text{srgb}}}{12.92} & \text{if } C_{\text{srgb}} \le 0.04045 \\ \left(\frac{C_{\text{srgb}} + 0.055}{1.055}\right)^{2.4} & \text{if } C_{\text{srgb}} > 0.04045 \end{cases}$$
2. Compute relative luminance:
   $$L = 0.2126 \cdot R_{\text{linear}} + 0.7152 \cdot G_{\text{linear}} + 0.0722 \cdot B_{\text{linear}}$$
3. Compute contrast ratio between lighter color $L_1$ and darker color $L_2$:
   $$\text{Ratio} = \frac{L_1 + 0.05}{L_2 + 0.05}$$

### Step 3: Non-Text Contrast Audit (SC 1.4.11)
Audit interactive UI elements:
- **Input Borders**: Form inputs with a border must have $\ge 3:1$ contrast against the adjacent canvas.
- **Focus Rings**: Focus indicator rings must have $\ge 3:1$ contrast against both the focused component background and the surrounding canvas.
- **Icons & Badges**: Standalone icons conveying meaning must have $\ge 3:1$ contrast against their immediate background.

### Step 4: Perceptual Validation with APCA
Evaluate readability with the Advanced Perceptual Contrast Algorithm (APCA):
- Body text ($16\text{px}$ regular): Minimum $L_c \ge 75$.
- Subtitles / bold UI elements ($14\text{px}$ bold): Minimum $L_c \ge 60$.
- Secondary hints / placeholders: Minimum $L_c \ge 45$.

### Step 5: Color Blindness & Use of Color Check (SC 1.4.1)
1. Simulate the UI under color vision deficiencies:
   - **Protanopia** (Long-wave cone deficiency / Red blindness).
   - **Deuteranopia** (Medium-wave cone deficiency / Green blindness).
   - **Tritanopia** (Short-wave cone deficiency / Blue blindness).
2. Enforce dual-coding: Verify that errors, success messages, and required fields are accompanied by an icon, text label, or underline—never color alone.

---

## 6. Decision Rules & Invariants
- **RULE 1 (Mandatory 4.5:1 for Normal Text)**: Any body text with contrast $< 4.5:1$ against its resolved background fails quality gates immediately.
- **RULE 2 (Mandatory 3:1 for Focus Rings & Borders)**: Interactive component boundaries and keyboard focus indicators must achieve $\ge 3:1$ contrast against adjacent surfaces.
- **RULE 3 (Dual Visual Cue Invariant)**: System states (error, warning, active, dirty) must provide a secondary non-color indicator (icon, badge, border pattern).
- **RULE 4 (Disabled State Exemption)**: Disabled elements are exempt from SC 1.4.3, but must remain distinguishable from active elements without causing legibility confusion.

---

## 7. Evidence Required
- **Mathematical Contrast Matrix**: Computed relative luminance and ratio for every foreground/background pair.
- **Interactive State Scorecard**: Contrast ratios verified across default, hover, focus, and active states.
- **Dual-Cue Verification**: Documented proof that all status indicators utilize text or icons in addition to color.

---

## 8. Output Contract
- Color Contrast Audit Report (`templates/contrast-report.md`).
- Pass/Fail scorecard covering all theme permutations (Light, Dark, High-Contrast).
- Remediation patches for non-compliant design tokens or CSS rules.

---

## 9. Stop Conditions
- 100% of text elements pass WCAG 2.2 AA ($4.5:1$ normal, $3:1$ large).
- 100% of focus rings and non-text elements pass $3:1$ contrast.
- Token budget exhausted.
- Blocked on external dependency.

---

## 10. Escalation Rules
- Escalate to lead designer if brand guidelines mandate a foreground/background pair that fails the $4.5:1$ threshold.
- Escalate if graphical data visualization charts cannot maintain $3:1$ contrast across $> 8$ categorical series without pattern fills.
