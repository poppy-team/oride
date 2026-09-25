# Zoom & Reflow Technical Guide (WCAG 2.2 SC 1.4.4 & SC 1.4.10)

## 1. Regulatory Standards and Normative Criteria

Accessibility guidelines mandate that web content and native interfaces remain fully readable, operable, and layout-stable when scaled by users with low vision or motor impairments:

1. **WCAG 2.2 Success Criterion 1.4.4: Resize Text (Level AA)**:
   - Except for captions and images of text, text can be resized without assistive technology up to 200 percent without loss of content or functionality.
   - Text scaling must not be clipped, truncated without user expansion, or obscured by adjacent fixed-positioned elements.
2. **WCAG 2.2 Success Criterion 1.4.10: Reflow (Level AA)**:
   - Content can be presented without loss of information or functionality, and without requiring scrolling in two dimensions for:
     - **Vertical scrolling content** at a width equivalent to **320 CSS pixels** (equivalent to 1280px viewport zoomed to 400%).
     - **Horizontal scrolling content** at a height equivalent to **256 CSS pixels** (equivalent to 1024px viewport zoomed to 400%).
   - Permitted exceptions: Content requiring two-dimensional layout for usage or meaning (data tables, maps, diagrams, video/canvas editing interfaces, code blocks with pre-formatted syntax).

---

## 2. CSS Architectural Strategies for Fluid Reflow

### 2.1 Viewport Configuration

The HTML `<meta name="viewport">` tag must never disable zooming or restrict user scaling:

```html
<!-- REQUIRED -->
<meta name="viewport" content="width=device-width, initial-scale=1">

<!-- STRICTLY PROHIBITED -->
<!-- <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1.0, user-scalable=no"> -->
```

### 2.2 Fluid Units and Typography Scaling

Never declare font sizes, element heights, or line heights using fixed absolute units (`px`, `pt`). Always use relative units (`rem`, `em`, `ch`, `%`):

```css
/* Fluid Typography with Clamp (Safe scaling within bounded constraints) */
:root {
  --font-base: 1rem; /* Respects user OS/browser base font size */
  --font-fluid-heading: clamp(1.5rem, 1rem + 2vw, 2.75rem);
  --line-height-body: 1.5;
  --line-height-heading: 1.25;
}

body {
  font-size: var(--font-base);
  line-height: var(--line-height-body);
}

h1 {
  font-size: var(--font-fluid-heading);
  line-height: var(--line-height-heading);
  word-break: break-word; /* Prevents long titles from causing horizontal overflow */
  overflow-wrap: anywhere;
}
```

### 2.3 Layout Reflow: Flexbox and Grid

Hardcoded multi-column widths (e.g., `width: 33.333%` or `min-width: 400px`) inevitably cause 2D scrolling at 320px viewport width. Use responsive wrapping and auto-fitting grid templates:

```css
/* Responsive Grid with minmax and min() clamping */
.responsive-grid {
  display: grid;
  /* Min column width collapses gracefully to 100% when parent is smaller than 280px */
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 280px), 1fr));
  gap: clamp(1rem, 2vw, 2rem);
  width: 100%;
}

/* Flexible Wrapping Layout */
.flexible-bar {
  display: flex;
  flex-wrap: wrap; /* Critical: allows children to stack rather than clip or push viewport */
  align-items: center;
  gap: 0.75rem;
}

.flexible-bar > .search-input {
  flex: 1 1 200px; /* Grow, shrink, basis of 200px (stacks when < 200px) */
  min-width: 0;   /* Prevents flex items from overflowing container min-content */
}
```

### 2.4 Handling Exception Elements (Data Tables, Code Blocks, Maps)

When content inherently requires two dimensions, wrap it in an accessible, keyboard-scrollable container:

```html
<div class="scrollable-region" 
     tabindex="0" 
     role="region" 
     aria-label="Financial summary table, horizontal scroll available">
  <table class="data-table">
    <!-- tabular data -->
  </table>
</div>
```

```css
.scrollable-region {
  max-width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  -webkit-overflow-scrolling: touch;
  border: 1px solid var(--border-neutral);
}

.scrollable-region:focus-visible {
  outline: 2px solid var(--color-focus-ring);
  outline-offset: 2px;
}
```

---

## 3. Automated Inspection and Headless Verification

Headless browsers (Playwright/Puppeteer) detect reflow violations by querying document geometry under simulated 320px and 256px viewports:

```javascript
// Verification snippet for Playwright / Puppeteer
test('Reflow compliance at 320px viewport', async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 1024 });
  await page.goto('/target-view');

  const horizontalScrollViolation = await page.evaluate(() => {
    const docWidth = document.documentElement.clientWidth;
    const scrollWidth = document.documentElement.scrollWidth;
    
    // Check if entire document causes horizontal scrolling
    if (scrollWidth > docWidth) {
      // Find the offending elements
      const elements = Array.from(document.querySelectorAll('*'));
      const offenders = elements.filter(el => {
        const rect = el.getBoundingClientRect();
        return rect.right > docWidth;
      }).map(el => ({
        tag: el.tagName,
        id: el.id,
        className: el.className,
        width: el.getBoundingClientRect().width
      }));
      return { scrollWidth, docWidth, offenders };
    }
    return null;
  });

  expect(horizontalScrollViolation).toBeNull();
});
```
