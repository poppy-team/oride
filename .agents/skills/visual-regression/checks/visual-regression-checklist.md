# Automated Visual Regression Testing (VRT) Checklist

## 1. Environment & Determinism
- [ ] **Standardized Container**: Tests execute within a fixed Docker container or CI runner with pinned FreeType/browser versions.
- [ ] **Deterministic Clock**: System time is frozen (`page.clock.setFixedTime()`) or mock dates injected.
- [ ] **Network Payloads Mocked**: Dynamic API data replaced with deterministic fixtures to prevent content shifting.
- [ ] **Device Pixel Ratio (DPR)**: Pinned explicitly to 1x or 2x in browser context settings.

## 2. Flakiness & Rendering Stabilization
- [ ] **Motion Disabled**: All CSS animations, transitions, and marquee scrolls forced to `0s !important`.
- [ ] **Font Readiness**: Awaits `document.fonts.ready` prior to snapshot capture.
- [ ] **Image Loading**: Awaits network idle or verifies naturalWidth/complete on all above-the-fold images.
- [ ] **Scroll Bar Stabilization**: Headless browser configured with `--hide-scrollbars` or custom CSS `scrollbar-width: none`.

## 3. Dynamic Masking & Tolerance
- [ ] **Dynamic Elements Masked**: Live clocks, fluctuating metrics, and user avatars covered with solid mask locators.
- [ ] **Calibrated Thresholds**: `threshold` (0.1 - 0.2) and `maxDiffPixelRatio` ($\le 0.001$) calibrated to reject styling bugs while tolerating minor antialiasing noise.
- [ ] **Failure Artifacts**: Failed tests output Baseline, Actual, and Diff PNGs with magenta mismatch highlighting.

## 4. Governance & PR Workflow
- [ ] **Baseline Commit Hygiene**: Baseline images tracked in Git or cloud bucket with explicit PR review gates.
- [ ] **Multi-Theme Coverage**: Snapshots captured for both Light and Dark themes across responsive breakpoints.
