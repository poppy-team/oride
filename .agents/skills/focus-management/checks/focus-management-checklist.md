# Focus Management Verification Checklist

## 1. Modal Dialog Invariants
- [ ] Initial focus moves inside modal immediately upon opening.
- [ ] Tab wraps from last focusable element to first focusable element.
- [ ] Shift+Tab wraps from first focusable element to last focusable element.
- [ ] Pressing `Escape` closes the modal.
- [ ] Focus is restored to the triggering element upon closing.
- [ ] Background content is marked `inert` or `aria-hidden="true"`.

## 2. Dynamic Removal & Recovery
- [ ] Deleting an element from a list shifts focus to the next or previous sibling.
- [ ] Focus never defaults or resets to `<body>` during dynamic DOM changes.

## 3. SPA Route Navigation
- [ ] Client-side route changes shift focus to the main heading (`<h1 tabIndex={-1}>`).
- [ ] Assistive technologies immediately announce the new page title.
