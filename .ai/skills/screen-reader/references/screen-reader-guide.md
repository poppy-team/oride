# Accessibility Tree & W3C AccName 1.2 Reference Guide

## 1. How Screen Readers Interact with Modern UIs
Screen readers do not inspect the visual DOM directly. They communicate with the operating system's native accessibility API:
- **macOS / iOS**: NSAccessibility / UIAccessibility
- **Windows**: UI Automation (UIA) / IAccessible2
- **Linux**: AT-SPI2
- **Android**: AccessibilityNodeInfo

The browser or native UI framework (e.g. AccessKit in Rust) compiles the UI into an internal graph known as the **Accessibility Tree**. Each node contains:
- **Role**: e.g., `ROLE_SYSTEM_PUSHBUTTON`
- **Name**: e.g., `"Submit Application"`
- **State**: e.g., `STATE_SYSTEM_FOCUSED`, `STATE_SYSTEM_EXPANDED`
- **Bounding Rect**: For magnifier focus tracking.

## 2. Accessible Name Computation (AccName 1.2) Precedence
When an assistive technology determines what text to announce for an element, it computes the Accessible Name following this deterministic algorithm:

1. **`aria-labelledby`**: If present, resolve each referenced ID in order, concatenate their visible inner text.
2. **`aria-label`**: If `aria-labelledby` is absent, use the exact string value of `aria-label`.
3. **Host Language Association**: e.g. `<label for="...">` associated with `<input>`.
4. **Subtree Text Content**: For elements allowing name from contents (like buttons, links, table headers), concatenate direct and descendant text nodes.
5. **Tooltips / Fallbacks**: `title` or `placeholder` attributes (only evaluated if steps 1-4 yield an empty string).

## 3. ARIA Live Region Best Practices
- Mount the live container on page load: `<div aria-live="polite" aria-atomic="true" class="sr-only" id="announcer"></div>`.
- Mutate the `textContent` of the container only when an announcement is needed.
- If the container is created dynamically and populated simultaneously, some screen readers (e.g. older NVDA versions) will miss the announcement because the DOM observer was not initialized.
