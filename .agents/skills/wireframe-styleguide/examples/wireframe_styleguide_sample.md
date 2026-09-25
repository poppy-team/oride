# Project Nova — Cloud Workspace Wireframe Styleguide

## 1. System Overview & Scope
- **Project**: Nova Cloud Engineering Console
- **Platform**: Responsive Web Application (Desktop & Tablet primary, Mobile secondary)
- **Breakpoints**: 320px (Mobile), 768px (Tablet), 1024px (Desktop), 1440px (Wide)
- **Status**: Approved Schematic Baseline

---

## 2. Graybox Design Token Definitions

### 2.1 CSS Custom Properties
```css
:root {
    /* Canvas & Surfaces */
    --wire-canvas:           #F8FAFC; /* Slate 50 */
    --wire-surface:          #FFFFFF; /* Pure White */
    --wire-neutral-subdued:  #F1F5F9; /* Slate 100 */

    /* Boundaries & Dividers */
    --wire-border-subtle:    #E2E8F0; /* Slate 200 (1.5:1 ratio) */
    --wire-border-strong:    #768296; /* Slate 450 (3.9:1 ratio - passes WCAG AA non-text >= 3:1) */

    /* Typography & Icons */
    --wire-text-primary:     #0F172A; /* Slate 900 (15.2:1 ratio - passes WCAG AAA) */
    --wire-text-secondary:   #64748B; /* Slate 500 (4.6:1 ratio - passes WCAG AA) */

    /* Functional Accents */
    --wire-focus:            #2563EB; /* Blue 600 (Focus ring) */
    --wire-error:            #DC2626; /* Red 600 (Validation alert) */

    /* Spacing Units (8pt Grid) */
    --wire-space-1:          4px;
    --wire-space-2:          8px;
    --wire-space-3:          12px;
    --wire-space-4:          16px;
    --wire-space-6:          24px;
    --wire-space-8:          32px;
    --wire-space-12:         48px;
}
```

### 2.2 Typography Scale
- `--wire-font-sans`: `-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`
- `--wire-font-mono`: `"SF Mono", "Cascadia Code", Menlo, monospace`
- `Display Title`: `32px` / line-height `40px` (Weight 700)
- `Heading 1`: `24px` / line-height `32px` (Weight 700)
- `Heading 2`: `20px` / line-height `28px` (Weight 600)
- `Heading 3`: `16px` / line-height `24px` (Weight 600)
- `Body Text`: `14px` / line-height `20px` (Weight 400)
- `Caption / Meta`: `12px` / line-height `16px` (Weight 500)
- `Monospace Code`: `13px` / line-height `18px` (Weight 400)

---

## 3. Schematic Placeholders & Component Symbols

### 3.1 Placeholders
- **Image Element**:
  ```text
  +---------------------------------------+
  | \                                   / |
  |   \       [ Image: Cluster Map ]    /   |
  |     \            (16:9)           /     |
  |   /                                 \   |
  | /                                     \ |
  +---------------------------------------+
  ```
- **User Avatar**: `( [UN] )` Circle diameter 36px with user initials.
- **Metric Sparkline**: `[ ∿∿∿∿∿∿ 99.8% ]` Line trend with value indicator.

---

## 4. Component State Matrix

| Component | Default | Hover | Focus-Visible | Active / Pressed | Disabled | Invalid / Error |
|---|---|---|---|---|---|---|
| **Deploy Button** | Fill: `#0F172A`, Text: `#FFFFFF`, min-h: 44px | Fill: `#334155` | 2px solid `#2563EB`, offset 2px | Fill: `#020617` | Fill: `#F1F5F9`, Text: `#64748B`, disabled | N/A |
| **Branch Input** | Fill: `#FFFFFF`, Border: 1px `#768296`, h: 44px | Border: 1px `#0F172A` | Border: 2px `#2563EB`, shadow ring | Border: 2px `#2563EB` | Fill: `#F1F5F9`, Border: 1px dashed `#E2E8F0` | Border: 2px `#DC2626`, inline error below |
| **Env Select** | Fill: `#FFFFFF`, Border: 1px `#768296`, glyph `[▼]`| Border: 1px `#0F172A` | 2px solid `#2563EB`, offset 2px | Dropdown menu open | Fill: `#F1F5F9`, cursor: not-allowed | Border: 2px `#DC2626` |
| **Auto-Scale Toggle**| Switch track: `#E2E8F0`, thumb: `#FFFFFF` | Track: `#768296` | 2px solid `#2563EB`, offset 2px | Track: `#0F172A` | Opacity: 0.4 | N/A |

---

## 5. Responsive Layout Specifications

| Viewport | Range | Columns | Gutter | Margin | Reflow & Component Rules |
|---|---|---|---|---|---|
| **Mobile** | `320px - 767px` | 4 | `16px` | `16px` | Single-column stack. Sidebar collapses to bottom navigation bar or top hamburger menu `[☰]`. Minimum tap target 44x44px. |
| **Tablet** | `768px - 1023px` | 8 | `20px` | `24px` | 2-column dashboard layout. Left navigation rail collapses to icon-only format. |
| **Desktop** | `1024px - 1439px`| 12 | `24px` | `32px` | 12-column fluid grid. Persistent left navigation sidebar (260px fixed). Top utility header. |
| **Wide** | `1440px+` | 12 | `24px` | Auto | Max container width 1280px, centered horizontally on `--wire-canvas`. |

---

## 6. View Schematic Annotation Ledger

### View: Workspace Deployment Console

```
+-----------------------------------------------------------------------------------+
|  [☰]  [ Project Nova ]                       [ Search resources... ] [1]  ( [JD] ) |
+-----------------------------------------------------------------------------------+
|  [ Navigation Rail ]  |  Deploy Cluster Environment                            [2]|
|  - Overviews          |  +------------------------------------------------------+ |
|  - Deployments [•]    |  | Target Environment: [ Production [▼] ]               | |
|  - Metrics            |  | Branch / Tag:       [ main [✎] ]                     | |
|  - Settings           |  | Replicas:           [ 5 [-] [+] ]                    | |
|                       |  | Auto-Scale:         [ (o) Enabled ]                  | |
|                       |  |                                                      | |
|                       |  | [ Cancel ]           [ Deploy Workload [🚀] ]   [3]  | |
|                       |  +------------------------------------------------------+ |
+-----------------------------------------------------------------------------------+
```

| Ref ID | Target Element | Trigger | Expected System Behavior | Validation & Errors | Responsive Behavior |
|---|---|---|---|---|---|
| `[1]` | Global Search | `onInput` | Debounces query (300ms) and presents overlay matching clusters | "No clusters found" displayed if empty | Collapses into search icon button on Mobile `<768px` |
| `[2]` | Environment Select | `onClick` / `KeyDown` | Expands option menu (`Staging`, `Production`, `Canary`) | Selecting Production requires confirmation modal if branch != `main` | Full-width dropdown on Mobile viewports |
| `[3]` | Deploy Workload CTA | `onClick` / `Enter` | Disables button, sets spinner `[⌛]`, issues POST `/api/v1/deploy` | If quota exceeded: renders `#DC2626` alert box above button | Stretches to 100% width button on Mobile viewports |
