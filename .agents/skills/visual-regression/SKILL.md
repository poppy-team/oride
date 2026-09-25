---
name: visual-regression
description: Automated pixel-level visual regression testing (VRT), Playwright snapshot integration, SSIM/pixelmatch diffing, dynamic masking, animation freezing, and deterministic CI rendering gates.
---

# Automated Visual Regression Testing (VRT) Contract

## 1. Purpose
Design, execute, and govern automated pixel-level Visual Regression Testing (VRT) suites across component libraries and full-page application views. Prevent unintended layout shifts, font regressions, color mutations, and responsive breakpoint bugs by enforcing deterministic headless environments, dynamic content masking, animation freezing, and strict pixelmatch/SSIM thresholds.

---

## 2. Use When
- Establishing CI visual gates for component design systems (Storybook, Playwright Test).
- Validating refactored CSS, Tailwind classes, or global design tokens against visual regressions.
- Auditing multi-theme rendering (Light, Dark, High-Contrast) across responsive viewports.
- Verifying cross-browser rendering parity (Chromium, WebKit, Firefox).

---

## 3. Do Not Use When
- Performing functional assertion testing (DOM text, API mocks, unit state logic) where Playwright standard assertions or Jest suffice.
- Auditing purely semantic accessibility or keyboard roving focus without visual rendering (use `keyboard-accessibility` or `screen-reader`).
- Rapid low-fidelity prototyping where designs change hourly.

---

## 4. Required Context
Before capturing or updating visual snapshot baselines, verify:
1. **Target Viewports & Device Matrix**: Desktop (1280x720), Tablet (768x1024), Mobile (375x667), and Device Pixel Ratio (DPR 1x vs 2x Retina).
2. **Theme Matrix**: Light mode, Dark mode, High-contrast mode.
3. **Dynamic Content Selectors**: Selectors for timestamps, user profile pictures, live counters, random transaction IDs, and ads.
4. **Target Rendering Engine**: Headless Chromium / WebKit running in a standardized Docker container to guarantee identical font rasterization.

---

## 5. Procedure

### Step 1: Flakiness Mitigation & Environment Stabilization
To prevent false-positive visual failures:
1. **Freeze Motion**: Inject CSS disabling all CSS transitions, animations, and smooth scrolling:
   ```css
   *, *::before, *::after {
     animation-duration: 0s !important;
     animation-delay: 0s !important;
     transition-duration: 0s !important;
     transition-delay: 0s !important;
   }
   ```
2. **Mock Clocks & Deterministic Data**: Freeze the system time via Playwright `page.clock.setFixedTime()` and mock all API network payloads.
3. **Await Font & Asset Readiness**: Wait explicitly for `document.fonts.ready` and all visible `<img>` tags to complete loading prior to taking the snapshot.

### Step 2: Dynamic Content Masking
1. Identify elements with unpredictable visual variation (e.g., live clocks, fluctuating metric charts, user avatars).
2. Pass locator masks to the snapshot capture command:
   ```typescript
   await expect(page).toHaveScreenshot('dashboard-view.png', {
     mask: [page.locator('[data-testid="live-timestamp"]'), page.locator('.user-avatar')],
     maskColor: '#ff00ff',
   });
   ```

### Step 3: Snapshot Capture & Comparison (Pixelmatch / SSIM)
1. Execute image comparison comparing the current capture against the committed golden baseline.
2. Configure threshold boundaries:
   - `threshold` (perceptual color delta): `0.1` to `0.2` (in CIELAB or YIQ color space).
   - `maxDiffPixelRatio`: Max acceptable ratio of mismatched pixels (default: $\le 0.001$, i.e. 0.1%).
3. Automatically generate a 3-way visual artifact upon failure:
   - **Baseline Image** (Approved Golden State).
   - **Current Image** (Test Execution State).
   - **Diff Image** (Mismatched pixels highlighted in high-visibility magenta `#ff00ff`).

### Step 4: Baseline Triage & Update Governance
1. If diffs represent legitimate intentional design modifications:
   - Review the diff in PR review.
   - Run `prumo test visual --update-snapshots` (or `npx playwright test --update-snapshots`).
   - Commit the updated baseline images with explicit PR approval.
2. If diffs represent unintended styling defects:
   - Reject the CI build and remediate CSS/component layout.

---

## 6. Decision Rules
1. **Zero Unfrozen Animations**: A snapshot test executed without animation freezing is an invalid, flaky test that must not merge into CI.
2. **Standardized CI Rasterization**: Baselines must be generated inside the exact same container image / Linux distribution as the CI runner to prevent macOS vs. Linux FreeType antialiasing discrepancies.
3. **No Unmasked Dynamic Timestamps**: Any test capturing a live date/time without a mask or frozen clock is a defect.
4. **Explicit PR Review for Baseline Mutations**: Golden image commits require visual sign-off from the design or frontend lead.

---

## 7. Evidence Required
- **VRT Test Run Log**: Playwright or pixelmatch test execution log passing with 0 mismatched pixels.
- **Diff Artifact Bundle**: When failures occur, baseline, actual, and diff PNG files published as CI artifacts.
- **Environment Hash**: Docker container image tag or OS/font configuration recorded with the baseline.

---

## 8. Output Contract
A production visual regression deliverable must contain:
1. Playwright test suite (`tests/visual/*.spec.ts`).
2. Playwright VRT configuration (`playwright.config.ts`).
3. Visual Regression Audit Report (`templates/visual-regression-report.md`).
4. Baseline image repository under version control.

---

## 9. Stop Conditions
- 100% of defined visual snapshot test cases match baseline within tolerance.
- Zero flaky false positives across 5 consecutive CI runs.
- Failure artifacts generate clear 3-way image diffs.

---

## 10. Escalation Rules
- Escalate to Design Lead if cross-browser rendering discrepancies (e.g. WebKit sub-pixel text rendering differences) cannot be reconciled within standard tolerance.
- Escalate to DevOps if CI container resource constraints cause headless browser screenshot timeouts.
