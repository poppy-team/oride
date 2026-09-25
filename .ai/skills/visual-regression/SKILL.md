---
name: visual-regression
description: Pixel-by-pixel snapshot testing, Playwright integration, dynamic content masking, deterministic font rendering, and cross-browser CI gates
---
# Automated Visual Regression Testing

## 1. Deterministic Environment Configuration
Run visual regression tests in standardized Docker containers to eliminate OS-level font antialiasing and sub-pixel rendering discrepancies across local and CI machines.

## 2. Playwright & Tooling Integration
Integrate automated snapshot testing using Playwright (expect(page).toHaveScreenshot()) or pixelmatch. Capture full-page and component-level screenshots across standard viewports (Mobile, Tablet, Desktop).

## 3. Dynamic Content Masking
Mask dynamic, fluctuating, or time-dependent UI elements (timestamps, user avatars, animated banners, random IDs) before snapshot capture using test locator masks.

## 4. Animation & Transition Freezing
Disable all CSS animations, smooth scrolling, and transitions during test runs (prefers-reduced-motion: reduce) to guarantee capture of static, fully settled layouts.

## 5. Font & Web Asset Loading Gates
Wait for document.fonts.ready and all critical images to finish loading prior to taking snapshots to prevent capturing half-rendered typography or layout shifts.

## 6. Configurable Thresholds & Tolerance
Set strict but realistic pixel-diff tolerance thresholds (e.g. maxDiffPixelRatio: 0.001). Distinguish between acceptable anti-aliasing variations and true visual regressions.

## 7. Side-by-Side Diff Artifact Reporting
Generate visual failure reports presenting Baseline, Current, and Diff images with highlighted pixel changes. Publish reports as CI artifacts on failed runs.

## 8. Approval & Baseline Update Workflows
Establish a controlled command workflow for updating baseline images (e.g. prumo test visual --update-snapshots) only when intentional UI updates are committed.

## 9. Component-Level Isolation Testing
Test visual snapshots of isolated component states in Storybook before testing complex end-to-end user journeys to isolate styling regressions rapidly.

## 10. Cross-Browser & DPI Coverage
Verify visual rendering across Chromium, Firefox, and WebKit rendering engines at both 1x and high-DPI (2x retina) pixel densities.
