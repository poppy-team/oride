# Visual Regression Testing (VRT) Strategies

Visual regression testing ensures that code changes do not inadvertently break the user interface visually.

## 1. Baselines and Comparisons
- **Baseline Image:** The approved visual state of a component or page.
- **Test Image:** The screenshot captured during the current test run.
- **Diff Image:** A generated artifact highlighting the pixel differences between the baseline and the test image in a bright color (usually red or magenta).

## 2. Managing False Positives
Visual tests are notoriously flaky due to environmental differences.
- **Anti-Aliasing:** Font rendering differs slightly between OS environments (e.g., macOS vs Linux CI). Use tools that support anti-aliasing detection.
- **Animations/Gifs:** Disable CSS animations, transitions, and SVG animations (`animation: none !important;`) before capturing screenshots.
- **Dynamic Content:** Mock API responses or mask dynamic UI elements (like dates, usernames, timestamps) with solid color blocks before capturing.

## 3. Thresholds
Do not enforce 0-pixel differences unless testing completely static, pixel-perfect environments.
- **Pixel Threshold:** Accept a small number of differing pixels.
- **Percentage Threshold:** Allow a minor percentage of the total image size to differ (e.g., `0.1%`).

## 4. Environment Consistency
Run tests in Docker containers to guarantee that the OS, browser version, and font rendering engine are identical across all developer machines and the CI pipeline.
