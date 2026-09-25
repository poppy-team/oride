# Programmatic Focus Management & Inertness Reference Guide

## 1. The Anatomy of Focus Trapping
Focus trapping is required whenever an overlay is modal. Without a focus trap, keyboard users tabbing through elements will blindly navigate into background links and form controls that are visually obscured behind a darkened backdrop.

### Native `<dialog>` Element vs Custom Overlays:
Whenever possible, prefer the HTML5 `<dialog>` element:
```html
<dialog id="confirm-modal">
  <form method="dialog">
    <p>Do you wish to proceed?</p>
    <button value="cancel">Cancel</button>
    <button value="confirm">Confirm</button>
  </form>
</dialog>
```
Invoking `dialog.showModal()` automatically:
- Traps keyboard focus within the dialog.
- Treats the rest of the document as `inert`.
- Dispatches `close` on `Escape`.

## 2. The HTML `inert` Attribute
The `inert` boolean attribute tells the browser to ignore user input events (clicks, focus) and hide the element from assistive technology trees:
```html
<div id="main-app" inert>
  <!-- Everything in here is completely untabbable and unclickable -->
</div>
<div id="active-modal" role="dialog">
  <!-- Interactive elements here work normally -->
</div>
```

## 3. Focus Restoration State Machine
```
[User on Trigger Button] -> (User clicks) -> [Cache activeElement]
                                                     |
                                                     v
                                         [Open Modal & Focus Inside]
                                                     |
                                         (User presses Escape/Closes)
                                                     |
                                                     v
                                         [activeElement.focus()]
```
If the trigger button was unmounted or replaced dynamically while the modal was open:
- Fallback to the nearest parent container or navigation landmark.
- Never allow `document.body` to inherit focus unassisted.
