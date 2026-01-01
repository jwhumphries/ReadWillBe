## 2025-12-23 - HTMX Loading Indicators with DaisyUI & Tailwind
**Learning:** Standard `.htmx-indicator` uses opacity for transitions, which reserves layout space and causes text shift even when hidden. To avoid this shift (especially in buttons), use `display: none` logic.
**Action:** Use the `group` pattern: add `class="group"` to the form (which receives `.htmx-request` during submission) and use `hidden group-[.htmx-request]:inline-block` on the spinner element. This ensures the spinner takes zero space when inactive.

## 2025-12-29 - Password Visibility Toggle Pattern
**Learning:** DaisyUI input groups with `label.input` wrapper are effective for adding icons inside inputs, but require restructuring from standard `input` elements.
**Action:** Use the `label.input` wrapper with `grow` on the input field and `btn-ghost` for the toggle button for consistent "Input with Button" patterns.

## 2025-01-01 - Input Placeholders for UX
**Learning:** Adding placeholders to input fields provides critical visual hints for expected formats, especially for authentication forms (e.g., email, password). This is a simple but high-impact micro-UX improvement.
**Action:** Ensure all input fields, especially in public-facing forms like Sign In/Up, have descriptive `placeholder` attributes.
