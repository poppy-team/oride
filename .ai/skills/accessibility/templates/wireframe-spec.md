# Accessible Screen Specification — Billing History

## Screen Purpose
Let an account owner find, inspect, and download invoice history without losing keyboard or screen-reader context.

## Primary Action
Filter the invoice list by date range and open one invoice.

## Regions and Landmarks
- `header`: product name, account menu, and page title.
- `nav[aria-label="Billing"]`: Current period, Invoices, Payment methods.
- `main`: invoice filters, result count, table, and pagination.
- `status[role="status"]`: loading and result-update announcements.

## Interaction Contract
- Date fields accept ISO 8601 calendar dates, such as `2026-09-01`, and expose format help.
- `Enter` submits filters; `Escape` clears the optional customer reference.
- Sort buttons announce `column, ascending/descending` through their accessible names.
- Row actions open in the same tab; focus returns to the triggering row button.

## State Matrix
| State | Visible content | Accessibility behavior |
|---|---|---|
| Loading | Table skeleton plus status message | `aria-busy=true`; focus stays on the submit button |
| Loaded | 18 invoices and date range | Sort and row actions become available |
| Empty | “No invoices match these filters” | Message references the active filters |
| Error | Retryable alert with correlation ID | Focus moves to alert; retry is keyboard reachable |
| Offline | Cached period plus offline notice | Stale-data text is explicit |

## Acceptance
- Normal text contrast is at least `4.5:1`; non-text focus and component indicators are at least `3:1`.
- At 400% zoom, the layout reflows without horizontal page scrolling.
- Automated checks have zero serious violations; NVDA and VoiceOver complete the primary action.
