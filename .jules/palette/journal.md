## 2025-12-23 - HTMX Loading Indicators with DaisyUI & Tailwind
**Learning:** Standard `.htmx-indicator` uses opacity for transitions, which reserves layout space and causes text shift even when hidden. To avoid this shift (especially in buttons), use `display: none` logic.
**Action:** Use the `group` pattern: add `class="group"` to the form (which receives `.htmx-request` during submission) and use `hidden group-[.htmx-request]:inline-block` on the spinner element. This ensures the spinner takes zero space when inactive.

## 2025-12-29 - Password Visibility Toggle Pattern
**Learning:** DaisyUI input groups with `label.input` wrapper are effective for adding icons inside inputs, but require restructuring from standard `input` elements.
**Action:** Use the `label.input` wrapper with `grow` on the input field and `btn-ghost` for the toggle button for consistent "Input with Button" patterns.

## 2025-12-31 - [Hero Pattern for Empty States]
**Learning:** Users often feel lost when confronting empty lists. Replacing generic alert banners with a "Hero" component (large icon, clear title, explanatory text, and primary CTA) significantly improves perceived value and guidability.
**Action:** When implementing empty states for core entities (like Plans), use the DaisyUI `hero` component with a centered layout and direct call-to-action button, ensuring the icon is visually distinct but not overwhelming (e.g., low opacity).

## 2025-01-01 - Input Placeholders for UX
**Learning:** Adding placeholders to input fields provides critical visual hints for expected formats, especially for authentication forms (e.g., email, password). This is a simple but high-impact micro-UX improvement.
**Action:** Ensure all input fields, especially in public-facing forms like Sign In/Up, have descriptive `placeholder` attributes.
## 2026-01-15 - Semantic Labels for Forms
**Learning:** Many form inputs were labeled with `span` elements instead of proper `<label>` tags, making them inaccessible to screen readers and preventing click-to-focus. Playwright's `get_by_label` fails on these, highlighting the issue.
**Action:** Always use `<label for="id">` for visible labels, and ensure the corresponding input has the matching `id`. For inputs without visible labels, use `aria-label`.
## 2026-01-19 - Tooltips for Icon-Only Buttons
**Learning:** Icon-only buttons (like "Remove" in a table row) can be ambiguous, even with `aria-label`. Sighted users benefit significantly from tooltips that clarify the action.
**Action:** Always wrap icon-only buttons in a DaisyUI `tooltip` component (e.g., `<div class="tooltip" data-tip="...">`). For destructive actions, use `tooltip-error`. For placement near edges (like right-aligned table cells), use `tooltip-left` or `tooltip-bottom` to prevent clipping.
