# Wireframe Styleguide & Schematic System — Engineering Reference Guide

## 1. Principles of Schematic & Graybox Design Systems

A wireframe styleguide establishes a rigorous visual grammar for low- and mid-fidelity interface planning. Its primary objective is to eliminate visual noise—such as decorative colors, brand typography, and unvetted photography—allowing engineering and product teams to focus purely on:
1. **Information Architecture**: Landmark hierarchy and structural relationships.
2. **Component State Completeness**: Edge cases, loading states, error boundaries, and interactive feedback.
3. **Responsive Geometry**: Layout reflow algorithms, grid columns, and viewport adaptations.
4. **Accessible Semantics**: Focus indicators, tab sequences, and touch targets.

---

## 2. Graybox Design Token Taxonomy

### 2.1 Color Ramp & Contrast Mathematics

All color tokens must be neutral grays, except for two functional signals: Focus Blue and Validation Error Red.

| Token Name | Hex Value | Purpose | Min Contrast Ratio |
|---|---|---|---|
| `--wire-canvas` | `#F8FAFC` | Page background canvas | N/A |
| `--wire-surface` | `#FFFFFF` | Card, container, and dialog background | $\ge 1.05:1$ vs Canvas |
| `--wire-neutral-subdued` | `#F1F5F9` | Table striping, secondary container fills | $\ge 1.1:1$ vs Surface |
| `--wire-border-subtle` | `#E2E8F0` | Dividers, subtle card boundaries | $\ge 1.5:1$ vs Surface |
| `--wire-border-strong` | `#768296` | Input outlines, active component boundaries | $\ge 3.0:1$ vs Surface (WCAG AA non-text) |
| `--wire-text-secondary` | `#64748B` | Captions, meta labels, placeholders | $\ge 4.5:1$ vs Surface (WCAG AA normal text) |
| `--wire-text-primary` | `#0F172A` | Headings, body copy, active labels | $\ge 14.0:1$ vs Surface |
| `--wire-focus` | `#2563EB` | 2px solid keyboard focus indicator | $\ge 3.0:1$ vs Surface & adjacent controls |
| `--wire-error` | `#DC2626` | Error borders, inline error text | $\ge 4.5:1$ vs Surface |

#### Contrast Ratio Calculation
The relative luminance $L$ of any sRGB color $(R, G, B)$ is defined as:
$$L = 0.2126 \cdot R_{\text{lin}} + 0.7152 \cdot G_{\text{lin}} + 0.0722 \cdot B_{\text{lin}}$$
where for each channel $C \in \{R, G, B\}$ normalized to $[0, 1]$:
$$C_{\text{lin}} = \begin{cases} \frac{C}{12.92} & \text{if } C \le 0.04045 \\ \left(\frac{C + 0.055}{1.055}\right)^{2.4} & \text{if } C > 0.04045 \end{cases}$$
The contrast ratio $CR$ between two colors with luminances $L_1$ (lighter) and $L_2$ (darker) is:
$$CR = \frac{L_1 + 0.05}{L_2 + 0.05}$$

### 2.2 Typography Scale

To maintain focus on content hierarchy, use a single system sans-serif font family with standard modular type scale (Minor Third, ratio $1.200$, or Major Second, ratio $1.125$):

| Scale Token | Font Size | Line Height | Font Weight | Semantic Usage |
|---|---|---|---|---|
| `--wire-text-display` | `32px` (`2.0rem`) | `40px` (`1.25`) | Bold (`700`) | Hero / Page Display Titles |
| `--wire-text-h1` | `24px` (`1.5rem`) | `32px` (`1.33`) | Bold (`700`) | Main Content Heading (`<h1>`) |
| `--wire-text-h2` | `20px` (`1.25rem`)| `28px` (`1.40`) | SemiBold (`600`) | Section Headings (`<h2>`) |
| `--wire-text-h3` | `16px` (`1.0rem`) | `24px` (`1.50`) | SemiBold (`600`) | Card / Sub-section Titles (`<h3>`) |
| `--wire-text-body` | `14px` (`0.875rem`)| `20px` (`1.43`)| Regular (`400`) | Standard Body Paragraphs |
| `--wire-text-caption`| `12px` (`0.75rem`) | `16px` (`1.33`)| Medium (`500`) | Captions, Timestamps, Meta tags |
| `--wire-text-code` | `13px` (`0.8125rem`)|`18px` (`1.38`)| Monospace (`400`) | Technical Annotations & Tokens |

### 2.3 Spacing & Layout Rhythm

Adhere strictly to an 8pt base grid for layout containers and 4pt for micro-component alignments:
$$\text{space}(n) = n \times 4\text{px} \quad \text{where } n \in \{1, 2, 3, 4, 6, 8, 12, 16, 24\}$$
- `--wire-space-1`: `4px` (micro padding, icon-text gap)
- `--wire-space-2`: `8px` (compact padding, input internal padding)
- `--wire-space-3`: `12px` (standard input horizontal padding)
- `--wire-space-4`: `16px` (card padding, list item gaps)
- `--wire-space-6`: `24px` (section gaps, grid gutters)
- `--wire-space-8`: `32px` (major container margins)
- `--wire-space-12`: `48px` (page section dividers)
- `--wire-space-16`: `64px` (landmark headers / hero offsets)

---

## 3. Schematic Symbol Conventions

### 3.1 Media & Imagery Placeholders
1. **Static Image**: Rectangular box with solid `--wire-border-subtle` border and diagonal internal cross (`X`). Labeled in center: `[Image: <Description> (<Aspect Ratio>)]`.
2. **Avatar**: Circle of diameter $32\text{px}$, $40\text{px}$, or $48\text{px}$ with `--wire-neutral-subdued` fill, containing user initials or a simple silhouette graphic.
3. **Video / Audio Player**: Rectangular box with centered triangle play symbol ($\blacktriangleright$) and horizontal scrub bar with duration indicator `[00:00 / 03:45]`.
4. **Data Visualizations**: Bounding box with dashed axis indicators ($X, Y$) and schematic bar heights or line paths with a legend placeholder `[Legend: A | B | C]`.

### 3.2 Iconography Representation
In low-fidelity wireframes, complex custom iconography causes distraction. Use standard ASCII or Unicode glyphs enclosed in square brackets:
- Navigation: `[☰]` Menu, `[←]` Back, `[→]` Forward, `[✕]` Close.
- Actions: `[+]` Create/Add, `[−]` Remove, `[✎]` Edit, `[🗑]` Delete, `[⬇]` Download.
- Status: `[✓]` Success, `[!]` Warning/Alert, `[ℹ]` Information, `[⌛]` Loading/Spinner.

---

## 4. Component State Matrix Specification

Every interactive component in the wireframe kit must define its visual representation across the standard 6 states:

```
+--------------------------------------------------------------------------+
|                       COMPONENT STATE LIFECYCLE                          |
|                                                                          |
|       [ Default ] ------( hover )-----> [ Hover ]                        |
|            |                               |                             |
|       ( keyboard )                     ( click )                         |
|            v                               v                             |
|   [ Focus-Visible ]                 [ Active/Pressed ]                   |
|            |                               |                             |
|       ( invalid )                     ( disabled )                       |
|            v                               v                             |
|     [ Error State ]                    [ Disabled ]                      |
+--------------------------------------------------------------------------+
```

| Component | Default | Hover | Focus-Visible | Active / Pressed | Disabled | Invalid / Error |
|---|---|---|---|---|---|---|
| **Primary Button** | Background: `--wire-text-primary`, Text: `--wire-surface`, Border: none | Background: `#334155` (10% lighter) | 2px solid `--wire-focus`, 2px offset | Background: `#020617` (inset press) | Background: `--wire-neutral-subdued`, Text: `--wire-text-secondary`, cursor: not-allowed | N/A (Buttons do not hold validation state) |
| **Secondary Button**| Background: `--wire-surface`, Border: 1px solid `--wire-border-strong` | Background: `--wire-neutral-subdued` | 2px solid `--wire-focus`, 2px offset | Background: `--wire-border-subtle` | Border: 1px dashed `--wire-border-subtle`, Text muted | N/A |
| **Text Input** | Background: `--wire-surface`, Border: 1px solid `--wire-border-strong` | Border: 1px solid `--wire-text-primary` | Border: 2px solid `--wire-focus`, shadow ring | Border: 2px solid `--wire-focus` | Background: `--wire-neutral-subdued`, Border: dashed | Border: 2px solid `--wire-error`, alert icon, error message |
| **Checkbox / Toggle**| Box: `--wire-surface`, Border: 1px solid `--wire-border-strong` | Border: 1px solid `--wire-text-primary` | 2px solid `--wire-focus`, 2px offset | Fill: `--wire-neutral-subdued` | Opacity: 0.4, cursor: not-allowed | Border: 2px solid `--wire-error` |

---

## 5. Wireframe Annotation Conventions

### 5.1 Callout Taxonomy & Badges
- Callout badges are numbered circles or brackets: `[1]`, `[2]`, `[3]`, rendered in bold monospace text with high contrast.
- Callouts must point directly to the interactive component or region boundary.

### 5.2 The Annotation Ledger Format
Every wireframe view must be accompanied by an Annotation Ledger structured as follows:

```markdown
### View: [View Name] — Annotation Ledger

| Ref ID | Target Component | Interaction Trigger | Expected System Behavior | Validation & Errors | Reflow / Mobile Behavior |
|---|---|---|---|---|---|
| [1] | Global Search Bar | `onInput` (debounce 250ms) | Triggers auto-complete dropdown showing top 5 matches | Display "No results found" if match count == 0 | Collapses to search icon button under 768px viewport |
| [2] | Save Changes CTA | `onClick` / `Enter` key | Sends POST payload; displays loading indicator on button | If form invalid: scrolls to first error field; focuses input | Full width (100% W) pinned to bottom sheet on mobile |
| [3] | Data Filter Drawer | `onClick` on Filter button | Opens modal drawer from right side; traps focus | Esc closes drawer without applying filters | Fullscreen overlay modal on mobile viewports |
```

---

## 6. Responsive Layout & Breakpoint Standards

| Viewport Category | Width Range | Grid Columns | Gutter Width | Page Margin | Structural Layout Behavior |
|---|---|---|---|---|---|
| **Mobile** | `320px - 767px` | 4 columns | `16px` | `16px` | Single-column stack; sidebars become off-canvas drawers; tables convert to stacked cards. |
| **Tablet** | `768px - 1023px` | 8 columns | `20px` | `24px` | 2-column grids; collapsible secondary navigation; compact metrics cards. |
| **Desktop** | `1024px - 1439px`| 12 columns | `24px` | `32px` | Full multi-column dashboard; persistent left navigation rail; multi-card grids. |
| **Wide Desktop** | `1440px+` | 12 columns (max $1280\text{px}$ container) | `24px` | Auto (centered) | Content container centered on canvas; extra margins on wide screens. |
