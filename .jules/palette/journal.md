## 2026-01-19 - Tooltips for Icon-Only Buttons
**Learning:** Icon-only buttons (like "Remove" in a table row) can be ambiguous, even with `aria-label`. Sighted users benefit significantly from tooltips that clarify the action.
**Action:** Always wrap icon-only buttons in a DaisyUI `tooltip` component (e.g., `<div class="tooltip" data-tip="...">`). For destructive actions, use `tooltip-error`. For placement near edges (like right-aligned table cells), use `tooltip-left` or `tooltip-bottom` to prevent clipping.
