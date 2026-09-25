# Color Contrast & Visual Legibility Audit Report

## 1. Audit Metadata
- **Target Application / View**: [e.g., Prumo Workspace Viewer / Settings Page]
- **Target Theme Mode**: [Light / Dark / High-Contrast]
- **Standard Evaluated**: WCAG 2.2 Level AA (Mandatory) & APCA
- **Audit Date**: YYYY-MM-DD

---

## 2. Text Elements Contrast Matrix (SC 1.4.3)
| Component / Text Selector | Text Size & Weight | Foreground Color | Background Color | Computed Ratio | Target Minimum | Status |
|---|---|---|---|---|---|---|
| Page Headings (`h1`, `h2`) | 24px Bold (Large) | `#0f172a` | `#f8fafc` | 17.06:1 | 3.0:1 | PASS |
| Body Content (`p`, `span`) | 16px Regular | `#1e293b` | `#ffffff` | 12.63:1 | 4.5:1 | PASS |
| Secondary Text / Meta | 14px Regular | `#475569` | `#ffffff` | 5.37:1 | 4.5:1 | PASS |
| Link Text (Default) | 16px Regular | `#1d4ed8` | `#ffffff` | 6.84:1 | 4.5:1 | PASS |
| Link Text (Hover) | 16px Regular | `#1e40af` | `#f1f5f9` | 7.92:1 | 4.5:1 | PASS |

---

## 3. Non-Text & Focus Indicators (SC 1.4.11)
| Component Element | Tested Border / Indicator | Adjacent Surface | Computed Ratio | Target Minimum | Status |
|---|---|---|---|---|---|
| Text Input Border (Default) | `#94a3b8` | `#ffffff` | 3.05:1 | 3.0:1 | PASS |
| Keyboard Focus Ring | `#2563eb` | `#ffffff` (Outer) | 4.62:1 | 3.0:1 | PASS |
| Keyboard Focus Ring | `#2563eb` | `#0f172a` (Inner) | 3.69:1 | 3.0:1 | PASS |
| Icon Buttons (Action Bar) | `#334155` | `#f8fafc` | 8.84:1 | 3.0:1 | PASS |

---

## 4. APCA Perceptual Legibility Analysis
- **Body Text APCA Score**: [e.g., $L_c 82$ (Pass: Target $\ge 75$)]
- **Secondary Caption APCA Score**: [e.g., $L_c 64$ (Pass: Target $\ge 60$)]
- **Dark Mode Polar Polarity Check**: [No halation / light-scattering reported]

---

## 5. Color Vision Deficiency (CVD) & Dual Visual Cue Audit (SC 1.4.1)
- [ ] Error messages display alert icon + red text (Pass: Not reliant on color alone).
- [ ] Form validation fields display clear textual error messages below input.
- [ ] Active tab indicated by high-contrast indicator bar + bold font weight.
- [ ] Tested under Protanopia, Deuteranopia, and Tritanopia color blindness filters.
