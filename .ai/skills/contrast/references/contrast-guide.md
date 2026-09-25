# Color Contrast, APCA & Perceptual Legibility Reference

## 1. WCAG 2.2 Relative Luminance & Contrast Formulation
The contrast ratio between two colors is calculated from their relative luminance values $L_1$ and $L_2$, where $L_1$ is the lighter color ($L_1 \ge L_2$):
$$\text{Contrast Ratio} = \frac{L_1 + 0.05}{L_2 + 0.05}$$

### 1.1 Linearizing sRGB Color Channels
Given an sRGB component $C_{\text{srgb}} \in [0, 1]$:
$$C_{\text{linear}} = \begin{cases} \frac{C_{\text{srgb}}}{12.92} & \text{if } C_{\text{srgb}} \le 0.04045 \\ \left(\frac{C_{\text{srgb}} + 0.055}{1.055}\right)^{2.4} & \text{if } C_{\text{srgb}} > 0.04045 \end{cases}$$

### 1.2 Computing Relative Luminance $L$
$$L = 0.2126 \cdot R_{\text{linear}} + 0.7152 \cdot G_{\text{linear}} + 0.0722 \cdot B_{\text{linear}}$$
- Pure Black ($#000000$): $L = 0.0$
- Pure White ($#ffffff$): $L = 1.0$
- Maximum possible contrast ratio: $\frac{1.0 + 0.05}{0.0 + 0.05} = 21:1$.

---

## 2. Alpha Compositing (Translucent Overlays)
When evaluating semi-transparent text or surfaces ($A_{\text{fg}} < 1.0$):
$$C_{\text{composite}} = C_{\text{fg}} \times A_{\text{fg}} + C_{\text{bg}} \times (1 - A_{\text{fg}})$$
The relative luminance must be calculated from the resulting composite color $C_{\text{composite}}$, not the raw foreground color.

---

## 3. APCA: Advanced Perceptual Contrast Algorithm
APCA (developed for WCAG 3.0) models human perceptual contrast $L_c$ considering spatial frequency and background polarity:

| Light on Dark vs Dark on Light | Human Perception Difference |
|---|---|
| Light text on Dark background | Halation / light scattering; requires thicker stroke weight |
| Dark text on Light background | Sharper perception; optimal for reading dense body copy |

### Recommended APCA $L_c$ Targets:
- $L_c \ge 90$: Thin / small text ($< 14\text{px}$).
- $L_c \ge 75$: Standard body text ($16\text{px}$ regular).
- $L_c \ge 60$: Headings, bold UI labels ($18\text{px}$ regular or $14\text{px}$ bold).
- $L_c \ge 45$: Secondary text, disabled elements, placeholder text.

---

## 4. Color Vision Deficiency (CVD) Simulation
Approximately 8% of males and 0.5% of females experience color vision deficiency:
- **Protanopia**: Absence of L-cones (long wavelength / red sensitive).
- **Deuteranopia**: Absence of M-cones (medium wavelength / green sensitive).
- **Tritanopia**: Absence of S-cones (short wavelength / blue sensitive).

### Simulation Pipeline:
1. Transform sRGB to linear LMS cone response space via matrix transformation.
2. Project LMS coordinates onto the dichromat confusion lines.
3. Transform back to sRGB.
4. If two status indicators (e.g. "Save" vs "Delete") have similar luminance under dichromat projection, a non-color differentiator (icon or explicit label) is strictly required by SC 1.4.1.
