# Motion Accessibility, Vestibular Disorders & Seizure Safety Reference

## 1. Vestibular Disorders & The Biology of Motion Sickness
The human vestibular system (located in the inner ear) controls balance, spatial orientation, and eye coordination. When a user observes large-scale visual motion on a screen while their body is physically stationary, the brain encounters a **visual-vestibular mismatch**:

```
Visual Perception:   [Rapid Screen Movement / Parallax / Full-page Zoom]
                                 VS
Vestibular Sensors:  [Body is Stationary at Rest (Zero G-Force Change)]
                                 =
Brain Response:      [Visual-Vestibular Conflict -> Vertigo, Severe Nausea, Migraines]
```

### High-Risk UI Patterns:
1. **Parallax Scrolling**: Background layers moving at different speeds than foreground text.
2. **Infinite Spinning Loaders**: Constant high-frequency rotation triggers dizziness.
3. **Full-Viewport Zooming**: Zooming in from a small card to take over the screen.
4. **Smooth Auto-Scroll**: Forcing viewport movement without direct manual touch/wheel input.

---

## 2. Photosensitive Epilepsy & Seizure Thresholds (SC 2.3.1)
Rapidly flashing light sequences can trigger seizures in people with photosensitive epilepsy.
- **The Critical Hazard Zone**: Flashes between $3\text{ Hz}$ (3 flashes per second) and $50\text{ Hz}$.
- **Area Threshold**: A flash is hazardous if it occupies more than $25\%$ of any 10-degree visual field.
- **Red Flash Rule**: Saturated red transitions ($R / (R + G + B) \ge 0.8$) transition into or out of darker colors at lower thresholds than neutral flashes and must be avoided completely.

---

## 3. CSS Progressive Motion Architecture
Rather than writing motion first and attempting to disable it later, use progressive enhancement:

### Pattern A: Opt-In Spatial Motion
```css
/* Base: Clean stationary state */
.card-drawer {
  opacity: 0;
  transition: opacity 150ms ease;
}
.card-drawer.open {
  opacity: 1;
}

/* Progressive: Add spatial translation only when user has NO motion restrictions */
@media (prefers-reduced-motion: no-preference) {
  .card-drawer {
    transform: translateY(20px);
    transition: opacity 250ms ease, transform 250ms cubic-bezier(0.16, 1, 0.3, 1);
  }
  .card-drawer.open {
    transform: translateY(0);
  }
}
```

### Pattern B: Universal Defensive Clamp
```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }
}
```

---

## 4. Native Desktop & Graphics Engine Hooks
### 4.1 Linux (GNOME / DBus)
```bash
gsettings get org.gnome.desktop.interface enable-animations
```
### 4.2 Windows (Win32 API)
```cpp
BOOL enableAnimations = TRUE;
SystemParametersInfo(SPI_GETCLIENTAREAANIMATION, 0, &enableAnimations, 0);
```
### 4.3 macOS (AppKit)
```swift
let reduceMotion = NSWorkspace.shared.accessibilityDisplayShouldReduceMotion
```
### 4.4 Web / JavaScript
```javascript
const prefersReduced = window.matchMedia("(prefers-reduced-motion: reduce)");
prefersReduced.addEventListener("change", (e) => {
  engine.setAnimationSmoothing(e.matches ? 0.0 : 0.85);
});
```
