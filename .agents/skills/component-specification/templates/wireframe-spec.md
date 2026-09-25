# Component Specification — DataTable

## 1. Responsibility
`DataTable` renders tabular records, manages sort intent, and exposes row selection. Pagination is owned by `TablePager`.

## 2. Props
| Prop | Type | Default | Contract |
|---|---|---|---|
| `columns` | `readonly Column<T>[]` | required | Stable `id`; label may be a string or accessible render function |
| `rows` | `readonly T[]` | `[]` | Caller owns collection updates |
| `sort` | `SortState | null` | `null` | Controlled sorting; `onSortChange` required |
| `selection` | `readonly string[]` | `[]` | Controlled selection; `onSelectionChange` required |
| `density` | `'compact' | 'comfortable'` | `'comfortable'` | Uses spacing tokens only |
| `emptyMessage` | `string` | `'No records found'` | Localized before rendering |

## 3. State Matrix
| State | Behavior | Focus and announcement |
|---|---|---|
| Loading | Preserve column headers, show row placeholders | Table is `aria-busy=true` |
| Empty | Render `emptyMessage` and optional action slot | Empty state is not focusable by default |
| Error | Render retry action with correlation ID | Focus moves only after explicit retry |
| Disabled | Prevent sort and selection mutation | Headers retain readable names |
| Dense | Reduce row padding to token value | Pointer target remains at least 24 px high |

## 4. Keyboard and Events
- `ArrowDown/ArrowUp` moves row focus without changing selection.
- `Enter` or `Space` toggles the focused row when selection is enabled.
- `Home/End` moves to first or last row.
- `onSortChange({ columnId, direction })` never emits a DOM node.

## 5. Acceptance
- Controlled and uncontrolled modes cannot be active simultaneously.
- Every interactive cell has an accessible name and visible focus.
- State fixtures cover default, loading, empty, error, disabled, and 200% zoom.
