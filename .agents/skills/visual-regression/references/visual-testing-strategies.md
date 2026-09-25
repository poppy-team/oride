# Visual Regression Testing Architecture & Comparison Mathematics

## 1. Image Comparison Metrics: Pixelmatch vs. SSIM vs. MSE

Automated visual testing evaluates image differences using three primary mathematical models:

### 1.1 Mean Squared Error (MSE)
Measures the cumulative squared error between two images $I_1$ and $I_2$ of dimension $M \times N$:
$$\text{MSE} = \frac{1}{M \cdot N} \sum_{i=1}^{M} \sum_{j=1}^{N} [I_1(i, j) - I_2(i, j)]^2$$
- *Limitation*: A 1-pixel layout shift across the whole image produces an enormous MSE error, despite the interface looking visually identical to human perception.

### 1.2 Structural Similarity Index Measure (SSIM)
SSIM models human visual perception by comparing local luminance, contrast, and structural patterns across matching window blocks $x$ and $y$:
$$\text{SSIM}(x, y) = \frac{(2\mu_x\mu_y + c_1)(2\sigma_{xy} + c_2)}{(\mu_x^2 + \mu_y^2 + c_1)(\sigma_x^2 + \sigma_y^2 + c_2)}$$
- $\mu_x, \mu_y$: Local mean pixel intensities.
- $\sigma_x, \sigma_y$: Local standard deviations (contrast).
- $\sigma_{xy}$: Covariance between window patches.
- *Value*: Ranges from -1 to 1 (1 indicates identical visual structure).

### 1.3 Perceptual Color Delta (Pixelmatch / YIQ / CIELAB)
Used by tools like Playwright and Pixelmatch. Calculates color distance $\Delta E$ in perceptual color space and applies anti-aliasing detection:
- Anti-aliasing pixels are adjacent to sharp high-contrast edges and blend between foreground and background.
- Pixelmatch detects if a mismatch is surrounded by higher/lower luminance neighbors; if so, it classifies the variation as anti-aliasing and suppresses false-positive failure.

---

## 2. Headless Browser Stabilization Flags

Operating system font rasterizers (FreeType on Linux, CoreText on macOS, DirectWrite on Windows) render sub-pixel antialiasing differently. To ensure 100% repeatable snapshots:

### Recommended Chromium Launch Flags
```javascript
const browser = await chromium.launch({
  args: [
    '--font-render-hinting=none',
    '--disable-skia-runtime-opts',
    '--disable-font-subpixel-positioning',
    '--disable-lcd-text',
    '--hide-scrollbars',
    '--force-color-profile=srgb',
  ]
});
```

---

## 3. Dynamic Masking and Clock Freezing in Playwright

```typescript
import { test, expect } from '@playwright/test';

test('Dashboard snapshot remains deterministic', async ({ page }) => {
  // 1. Freeze system clock to a fixed epoch
  await page.clock.setFixedTime(new Date('2026-09-23T12:00:00Z'));

  // 2. Navigate and wait for network/fonts
  await page.goto('/dashboard');
  await page.evaluate(() => document.fonts.ready);

  // 3. Capture screenshot with dynamic elements masked
  await expect(page).toHaveScreenshot('dashboard-baseline.png', {
    mask: [
      page.locator('[data-testid="live-activity-stream"]'),
      page.locator('.user-avatar-image')
    ],
    maxDiffPixelRatio: 0.001, // Permit up to 0.1% pixel variance (anti-aliasing)
    animations: 'disabled'
  });
});
```
