---
name: focus-management
description: Programmatic focus management, modal focus traps, focus restoration on dismissal, initial focus targeting, SPA route transition announcements, and DOM element removal focus recovery.
---

# Focus Management & Transition Architecture

## 1. Title and Description
**Focus Management & Transition Architecture (`focus-management`)**
Governs programmatic focus transitions, modal dialog focus traps, initial focus positioning, focus restoration upon overlay dismissal, and safe focus recovery during dynamic DOM mutations and SPA route changes.

## 2. Purpose
Prevent the dreaded "focus lost to document body" trap that disorients keyboard and screen reader users when modals open, dialogs close, content is deleted, or client-side routes transition.

## 3. Prerequisites
- UI component hierarchy managing dynamic overlays, dialogs, drawers, or single-page application (SPA) routing.
- Familiarity with the W3C ARIA Authoring Practices Guide (APG) Dialog pattern.

## 4. Inputs
- Modal dialogs, drawer components, popovers, dynamic list managers, or router hooks.
- Focus transition graph and modal lifecycle specifications.

## 5. Outputs
- Robust focus management code incorporating focus trapping, restoration, and transition announcements.
- Test suites asserting focus position before, during, and after UI mutations.
- Audit report conforming to `templates/focus-management-report.md`.

## 6. Execution Steps

### Step 1: The Three Invariants of Modal Overlays
Every modal dialog, drawer, or lightbox must adhere to three mandatory focus rules:
1. **Initial Focus Targeting**:
   - Upon mounting/opening, focus must immediately be moved inside the overlay.
   - Target the first interactive input field (e.g. text input in a form) or the primary action.
   - If the modal is a destructive confirmation dialog, focus the **Cancel** button by default to prevent accidental destruction on Enter.
2. **Focus Trapping (Constrained Tabbing)**:
   - When focus is on the last focusable element in the modal and user presses `Tab`, wrap focus around to the first focusable element.
   - When focus is on the first element and user presses `Shift+Tab`, wrap focus to the last focusable element.
   - Native HTML alternative: Use `<dialog>` with `.showModal()`, which manages focus trapping and backdrop inertness natively.
3. **Focus Restoration**:
   - Before opening the overlay, cache the currently active element: `const previousTrigger = document.activeElement;`.
   - When the overlay closes (via `Escape`, close button, or submission), restore focus: `previousTrigger?.focus();`.

### Step 2: Focus Recovery on Dynamic DOM Removal
When an item is deleted from a dynamic list or table:
1. Never allow the browser to reset focus to `document.body`.
2. **Focus Fallback Hierarchy**:
   - Focus the next sibling element in the list.
   - If the deleted item was the last element, focus the previous sibling.
   - If the list is now empty, focus the parent container or the "Add Item" button.

### Step 3: Single-Page Application (SPA) Route Transitions
In client-side single-page applications, route transitions do not trigger a browser page reload:
1. Upon route change, shift programmatic focus to the primary view heading (`<h1 tabIndex={-1}>`), or to a dedicated skip-navigation landmark.
2. Ensure the heading has `outline: none` but triggers assistive technology to announce the new page title immediately.

### Step 4: Non-Modal Popovers & Floating Tooltips
- For non-modal floating tooltips or info popovers: **do not trap focus**.
- Allow normal tab flow past the popover, but ensure pressing `Escape` dismisses the popover and leaves focus on the trigger element.

## 7. Verification
```bash
# 1. Automated focus lifecycle test
npm test -- src/tests/focus-trap.test.ts

# 2. Verify modal restoration in Playwright / Cypress
npx playwright test tests/modal-focus.spec.ts
```

## 8. Fallbacks & Failure Recovery
| Failure Class | Root Cause | Deterministic Remediation |
|---|---|---|
| Focus Drops to Body on Close | Trigger button was unmounted while modal was open | Fallback to focusing the main container or document header: `(trigger?.isConnected ? trigger : fallbackElement).focus()`. |
| Focus Escapes Modal via Shift+Tab | Reverse wrap condition failed | Ensure event listener intercepts `e.shiftKey && e.key === 'Tab'` on the first focusable element. |
| Background Elements Clickable Behind Modal | Modal does not isolate background inertness | Add `inert` attribute to outer app root while modal is open: `document.getElementById('root')?.setAttribute('inert', '')`. |

## 9. Constraints
- **Zero focus loss**: An interface must never lose focus to `<body>` during standard interactive workflows.
- **Always restore focus**: Closing a temporary overlay must return focus to the trigger that summoned it.

## 10. Examples

### Anti-Pattern: Modal Without Focus Trap or Restoration
```tsx
// BAD: Focus remains on background, Tab escapes into hidden content, close loses focus
export const BadModal = ({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) => {
  if (!isOpen) return null;
  return (
    <div className="modal-backdrop">
      <div className="modal-content">
        <h2>Confirm Action</h2>
        <button onClick={onClose}>Close</button>
      </div>
    </div>
  );
};
```

### Idiomatic Pattern: Production Focus Trap and Restoration
```tsx
// GOOD: Full initial focus, focus trap, escape listener, and focus restoration
import React, { useEffect, useRef } from 'react';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
}

export const AccessibleModal: React.FC<ModalProps> = ({ isOpen, onClose, title, children }) => {
  const modalRef = useRef<HTMLDivElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!isOpen) return;

    // 1. Cache the element that triggered the modal
    previousFocusRef.current = document.activeElement as HTMLElement;

    // 2. Focus the first interactive element inside modal
    const focusable = modalRef.current?.querySelectorAll<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    if (focusable && focusable.length > 0) {
        focusable[0].focus();
    }

    // 3. Keydown listener for Escape and Tab wrapping
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
        return;
      }

      if (e.key === 'Tab' && focusable && focusable.length > 0) {
        const first = focusable[0];
        const last = focusable[focusable.length - 1];

        if (e.shiftKey && document.activeElement === first) {
          e.preventDefault();
          last.focus();
        } else if (!e.shiftKey && document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      // 4. Restore focus on unmount/close
      previousFocusRef.current?.focus();
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" role="dialog" aria-modal="true" aria-labelledby="modal-title">
      <div ref={modalRef} className="w-full max-w-md p-6 bg-white rounded-lg shadow-xl">
        <h2 id="modal-title" className="text-lg font-bold mb-4">{title}</h2>
        {children}
      </div>
    </div>
  );
};
```
