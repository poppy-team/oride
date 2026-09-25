# Screen Reader & Accessibility Tree Verification Checklist

## 1. Accessible Names (AccName 1.2)
- [ ] Every `<button>` has visible text or an explicit `aria-label` / `aria-labelledby`.
- [ ] All icon-only buttons (`<button><svg/></button>`) have an `aria-label`.
- [ ] Every form `<input>`, `<select>`, `<textarea>` has an associated `<label for="id">`.
- [ ] Informational images have meaningful `alt="..."`; decorative images have `alt=""` and `aria-hidden="true"`.
- [ ] Inline SVGs used for decorative purposes are marked `aria-hidden="true" focusable="false"`.

## 2. Roles & Landmarks
- [ ] High-level landmarks present (`<header>`, `<nav>`, `<main>`, `<footer>`).
- [ ] No fake custom roles (`role="custom-box"` is invalid).
- [ ] No redundant ARIA markup on semantic elements (`<button role="button">`).

## 3. Dynamic Updates & Live Regions
- [ ] Asynchronous updates (toast notifications, search results count) announce via `aria-live="polite"`.
- [ ] Critical errors announce via `role="alert"` or `aria-live="assertive"`.
- [ ] Form error messages are programmatically linked to inputs via `aria-describedby`.
