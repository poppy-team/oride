# Motion Accessibility & Vestibular Safety (WCAG 2.2 SC 2.3.1, 2.3.2, 2.3.3)

## 1. Purpose
Design, implement, and audit motion effects, transitions, and animations across web and native desktop applications to prevent vestibular disorders, motion sickness, and photosensitive epileptic seizures. This skill mandates strict compliance with **WCAG 2.2 SC 2.3.1 (Three Flashes), SC 2.3.2 (No Flashes), and SC 2.3.3 (Animation from Interactions)**, universal support for `@media (prefers-reduced-motion: reduce)` and native OS reduced-motion flags, substitution of spatial movement with subtle opacity fades, and eradication of non-essential auto-playing animations.

---

## 2. Use When
- Developing UI animations, page route transitions, modal appearances, micro-interactions, or toast notifications.
- Implementing parallax scrolling effects, smooth scroll behaviors, or canvas/3D camera transitions.
- Building native desktop (egui, Slint, Qt, Electron) or web interfaces that consume system accessibility preferences.
- Auditing visual effects to eliminate rapid flashing (> 3 flashes/second) or abrupt visual disorientation.
- Operating in mode(s): `implementation`, `review`, `testing`, `audit`.

---

## 3. Do Not Use When
- Implementing static color palettes or verifying contrast ratios without animation (use `contrast`).
- Managing keyboard focus transitions without visual motion considerations (use `focus-management`).
- Developing offline batch asset compression with no runtime display rendering.

---

## 4. Required Context
Before implementing or modifying motion behavior, verify:
- **Motion Classification**: Essential motion (e.g. drawing canvas, interactive timeline scrubbing) vs Non-essential motion (e.g. decorative bounces, parallax backgrounds, page sliding transitions).
- **Target Platform Media Queries**: CSS `@media (prefers-reduced-motion: reduce)` for web, or native platform APIs (macOS `NSWorkspace.accessibilityDisplayShouldReduceMotion`, Windows `SPI_GETCLIENTAREAANIMATION`, Linux `org.gnome.desktop.interface.enable-animations`).
- **Safety Thresholds**: Zero flashes between 3 Hz and 50 Hz; maximum allowable transition duration under reduced motion $\le 50\text{ms}$ (instant or fade).

---

## 5. Procedure

```
[Animation / Transition Design]
         |
         v
[1. Flashing & Seizure Safety Check] -> Ensure flash rate is strictly < 3 Hz
         |
         v
[2. Reduced-Motion Media Query] ------> Wrap spatial transforms in prefers-reduced-motion
         |
         v
[3. Motion Substitution Pattern] -----> Replace transform/scale with simple opacity fade
         |
         v
[4. JS / Canvas Engine Integration] --> Bind window.matchMedia or native OS signals
         |
         v
[5. Automated Motion Audit Gauntlet] -> Headless Playwright / AST scanner verification
```

### Step 1: Flashing & Seizure Safety (SC 2.3.1)
- Verify that no UI element (spinners, warning strobes, video clips, banner ads) flashes more than 3 times in any 1-second window.
- High-saturation red flashes (`#ff0000`) are strictly banned at any frequency above 1 Hz.

### Step 2: CSS Universal Reduced-Motion Reset
Every web project must declare a universal defensive fallback:
```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}
```

### Step 3: Progressive Motion Design (Motion vs Opacity)
Rather than abruptly disabling all feedback, replace vestibular-triggering spatial displacement (`transform: translate`, `scale`, parallax) with non-spatial opacity fades:
```css
/* Default: Full spatial slide-in */
.modal-enter {
  opacity: 0;
  transform: translateY(30px) scale(0.95);
  transition: opacity 250ms ease, transform 250ms ease;
}

/* Reduced Motion: Clean, stationary opacity fade */
@media (prefers-reduced-motion: reduce) {
  .modal-enter {
    opacity: 0;
    transform: none;
    transition: opacity 100ms ease;
  }
}
```

### Step 4: JavaScript & Native Canvas Synchronization
In JavaScript, canvas, or 3D viewports (Three.js, WebGPU):
```javascript
export function setupMotionController(renderEngine) {
  const mediaQuery = window.matchMedia("(prefers-reduced-motion: reduce)");

  function updateMotionPolicy() {
    renderEngine.setReducedMotion(mediaQuery.matches);
  }

  mediaQuery.addEventListener("change", updateMotionPolicy);
  updateMotionPolicy();
}
```
When `mediaQuery.matches` is true:
- Camera smoothing and lerping are disabled (camera snaps directly to target).
- Ambient particle velocity is set to 0.
- Auto-playing looping animations are paused.

### Step 5: Native Desktop Engine Integration
In native Rust / C++ GUI frameworks (egui, Slint):
- Read the platform animation policy flag.
- Set UI animation times and transition durations to zero when enabled:
  ```rust
  let reduce_motion = ctx.os().system_theme.as_ref().map_or(false, |t| t.reduce_motion);
  if reduce_motion {
      // Disable cursor blinking, smooth tab sliding, and window fade-ins
  }
  ```

---

## 6. Decision Rules & Invariants
- **RULE 1 (Zero Flashing > 3Hz)**: Any component displaying content that flashes or strobes between 3 and 50 Hz fails quality gates immediately.
- **RULE 2 (Universal Reduced-Motion Handling)**: Every CSS animation or transition must either be scoped within `@media (prefers-reduced-motion: no-preference)` or provide an explicit `@media (prefers-reduced-motion: reduce)` override.
- **RULE 3 (No Parallax without Opt-In)**: Parallax background scrolling is strictly forbidden unless the user explicitly opts in via an accessible application toggle and system reduced-motion is inactive.
- **RULE 4 (Scroll Behavior Auto)**: `scroll-behavior: smooth` must always revert to `auto` when reduced-motion is requested, preventing scrolling-induced vertigo.

---

## 7. Evidence Required
- **Static CSS / JS Audit**: AST verification showing zero unbounded `@keyframes` lacking reduced-motion media query guards.
- **Dynamic Playwright Evidence**: Automated test runs with `emulateMedia({ reducedMotion: 'reduce' })` confirming animation durations are $\le 0.01\text{ms}$.
- **Flashing Frequency Validation**: Mathematical frequency analysis confirming all animated cycles remain below 3 Hz.

---

## 8. Output Contract
- Motion-safe stylesheets and transition definitions.
- JavaScript motion controller modules (`motion-controller.js`).
- Motion Accessibility Audit Report (`templates/motion-accessibility-report.md`).

---

## 9. Stop Conditions
- All animations respect `prefers-reduced-motion: reduce` across automated test gauntlet.
- Zero flashes above 3 Hz detected.
- Token budget exhausted.
- Blocked on external dependency.

---

## 10. Escalation Rules
- Escalate to product lead if a third-party embedded widget or video player lacks API hooks to disable auto-play motion.
- Escalate if hardware platform lacks OS accessibility query APIs for native reduced-motion flags.
