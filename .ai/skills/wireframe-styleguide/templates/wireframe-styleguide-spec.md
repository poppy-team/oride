# Wireframe Styleguide & Schematic System Specification Template

## 1. System Overview & Scope
- **Project Name**: [Project / Platform Identifier]
- **Target Platforms**: [Web Desktop, Responsive Mobile Web, Native iOS/Android]
- **Target Viewport Range**: [Minimum 320px to Maximum 1920px]
- **Author / UX Lead**: [Name / Identifier]
- **Revision & Date**: [Version 1.0, YYYY-MM-DD]

---

## 2. Graybox Design Token Architecture

### 2.1 Neutral Graybox Palette
```css
:root {
    /* Backgrounds & Canvas */
    --wire-canvas:           #F8FAFC;
    --wire-surface:          #FFFFFF;
    --wire-neutral-subdued:  #F1F5F9;

    /* Boundaries & Borders */
    --wire-border-subtle:    #E2E8F0;
    --wire-border-strong:    #768296;

    /* Content & Typography */
    --wire-text-primary:     #0F172A;
    --wire-text-secondary:   #64748B;

    /* Functional Accents */
    --wire-focus:            #2563EB;
    --wire-error:            #DC2626;
}
```

### 2.2 Typography Scale & Rhythm
- **Primary Typeface**: `-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`
- **Annotation Typeface**: `ui-monospace, "SF Mono", "Cascadia Code", monospace`
- **Scale Definition**:
  - `Display`: `32px` / line-height `40px` / Bold (700)
  - `Heading 1`: `24px` / line-height `32px` / Bold (700)
  - `Heading 2`: `20px` / line-height `28px` / SemiBold (600)
  - `Heading 3`: `16px` / line-height `24px` / SemiBold (600)
  - `Body`: `14px` / line-height `20px` / Regular (400)
  - `Caption`: `12px` / line-height `16px` / Medium (500)
  - `Code / Annotation`: `13px` / line-height `18px` / Regular (400)

### 2.3 Spacing Scale (8pt System)
- `--wire-space-1`: `4px`
- `--wire-space-2`: `8px`
- `--wire-space-3`: `12px`
- `--wire-space-4`: `16px`
- `--wire-space-6`: `24px`
- `--wire-space-8`: `32px`
- `--wire-space-12`: `48px`
- `--wire-space-16`: `64px`

---

## 3. Schematic Placeholders & Symbol Conventions
- **Image Bounding Box**: Solid `--wire-border-subtle`, diagonal cross `X`, centered label `[Image: <Description> (<Ratio>)]`.
- **Avatar**: Circle diameter ($32\text{px}$ / $40\text{px}$), neutral fill `--wire-neutral-subdued`, user initials.
- **Video / Audio**: Rectangular box with triangle play glyph `[ ▶ ]` and scrub bar.
- **Data Chart**: Dashed axis container with bar/line placeholder silhouette and legend block.
- **Iconography**: Standard text-enclosed glyphs `[+]`, `[−]`, `[✕]`, `[←]`, `[→]`, `[☰]`, `[✓]`, `[!]`.

---

## 4. Component State Matrix

| Component | Default | Hover | Focus-Visible | Active / Pressed | Disabled | Invalid / Error |
|---|---|---|---|---|---|---|
| **Primary Action** | Fill: `--wire-text-primary`, Text: `--wire-surface` | Fill: `#334155` | 2px solid `--wire-focus`, 2px offset | Inset active shade | Fill: `--wire-neutral-subdued`, text muted | N/A |
| **Secondary Action** | Fill: `--wire-surface`, Border: `--wire-border-strong` | Fill: `--wire-neutral-subdued` | 2px solid `--wire-focus`, 2px offset | Inset active shade | Dashed border, text muted | N/A |
| **Input Field** | Fill: `--wire-surface`, Border: `--wire-border-strong` | Border: `--wire-text-primary` | 2px solid `--wire-focus`, focus shadow | 2px solid `--wire-focus` | Fill: `--wire-neutral-subdued`, border dashed | 2px solid `--wire-error`, error text |
| **Select / Dropdown** | Fill: `--wire-surface`, Chevron: `[▼]` | Border: `--wire-text-primary` | 2px solid `--wire-focus`, 2px offset | Options panel opened | Fill: `--wire-neutral-subdued` | 2px solid `--wire-error` |
| **Checkbox / Toggle**| Box: `--wire-surface`, Border: `--wire-border-strong` | Border: `--wire-text-primary` | 2px solid `--wire-focus`, 2px offset | Fill: `--wire-neutral-subdued` | Opacity: 0.4, unselectable | Border: 2px solid `--wire-error` |

---

## 5. Responsive Grid & Breakpoint Rules

| Viewport | Range | Columns | Gutter | Margin | Reflow Rules |
|---|---|---|---|---|---|
| **Mobile** | `320px - 767px` | 4 | `16px` | `16px` | Single-column linear reflow; off-canvas navigation drawer; min 44px touch targets. |
| **Tablet** | `768px - 1023px` | 8 | `20px` | `24px` | 2-column card layouts; compact tabular data; collapsible sidebar. |
| **Desktop** | `1024px - 1439px`| 12 | `24px` | `32px` | Multi-column layouts; persistent left sidebar navigation; full data grids. |
| **Wide** | `1440px+` | 12 | `24px` | Auto | Centered content container (max-width `1280px`). |

---

## 6. View Schematic Annotation Ledger

| Ref ID | Target Element | Trigger | Expected System Behavior | Error & Edge Conditions | Responsive Reflow |
|---|---|---|---|---|---|
| `[1]` | `[Element Name]` | `[e.g., Click, Keypress]` | `[Description of action]` | `[Fallback or error state]` | `[Reflow behavior]` |
| `[2]` | `[Element Name]` | `[e.g., Change]` | `[Description of action]` | `[Fallback or error state]` | `[Reflow behavior]` |

---

## 7. Verification & Sign-Off Checklist
- [ ] All colors derive from `--wire-*` tokens; zero arbitrary hex values.
- [ ] Text tokens satisfy WCAG 2.1 AA ($\ge 4.5:1$).
- [ ] Component boundaries satisfy $\ge 3:1$ contrast.
- [ ] All interactive elements document 6 states and $\ge 44 \times 44\text{px}$ touch targets.
- [ ] Verification script passes without errors.
