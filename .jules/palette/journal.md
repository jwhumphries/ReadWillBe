## 2025-12-23 - HTMX Loading Indicators with DaisyUI & Tailwind
**Learning:** Standard `.htmx-indicator` uses opacity for transitions, which reserves layout space and causes text shift even when hidden. To avoid this shift (especially in buttons), use `display: none` logic.
**Action:** Use the `group` pattern: add `class="group"` to the form (which receives `.htmx-request` during submission) and use `hidden group-[.htmx-request]:inline-block` on the spinner element. This ensures the spinner takes zero space when inactive.
