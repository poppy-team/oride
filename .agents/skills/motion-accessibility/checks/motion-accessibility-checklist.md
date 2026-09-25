# Motion Accessibility & Vestibular Safety Invariants Checklist

## 1. Seizure & Flashing Prevention (WCAG 2.2 SC 2.3.1 & 2.3.2)
- [ ] No screen content flashes or strobes more than 3 times in any 1-second period.
- [ ] Saturated red flashes (`#ff0000`) are completely absent from visual effects.
- [ ] High-contrast animated patterns (e.g. zebra stripes, rapid alternating checkerboards) are eliminated.

## 2. Reduced-Motion Media Query Coverage (WCAG 2.2 SC 2.3.3)
- [ ] Universal defensive CSS reset is declared for `@media (prefers-reduced-motion: reduce)`.
- [ ] Spatial translations (`translate`, `scale`, `rotate`) are eliminated or replaced with subtle stationary opacity fades.
- [ ] `scroll-behavior: smooth` reverts to `scroll-behavior: auto` under reduced-motion.
- [ ] Transition and animation durations under reduced-motion are clamped to $\le 0.01\text{ms}$ or subtle $\le 100\text{ms}$ fades.

## 3. Vestibular Disorder & Parallax Defense
- [ ] Parallax scrolling layers are disabled when reduced-motion is requested.
- [ ] Full-screen zoom transitions (e.g. zooming from card into full page) are disabled or replaced with instant modal appearance.
- [ ] Background video or looping decorative canvas animations provide an accessible Pause/Play control.

## 4. JavaScript, Canvas & 3D Synchronization
- [ ] `window.matchMedia('(prefers-reduced-motion: reduce)')` listener is registered in interactive WebGL / Three.js / WebGPU engines.
- [ ] Camera lerping / spring smoothing snaps directly to target positions under reduced-motion.
- [ ] Particle systems reduce or zero out velocity vectors when reduced-motion is active.

## 5. Native Desktop Accessibility Invariants
- [ ] Desktop applications (egui, Slint, Qt, Electron) read operating system animation toggles at startup and listen to OS event changes.
- [ ] Cursor blinking, smooth tab switching, and window open/close animations conform to OS reduced-motion state.
