# Frontend Code Review - ReadWillBe

## Summary

| Category | Count | Severity |
|----------|-------|----------|
| Accessibility Issues | 6 | High |
| HTML Structure Issues | 3 | Medium |
| JavaScript Issues | 7 | Medium |
| DaisyUI Pattern Issues | 5 | Low |
| Missing Features/UX | 3 | Medium |
| CSS Issues | 4 | Low |
| State Management | 2 | Medium |
| Performance | 2 | Medium |
| Layout/UX | 3 | Low |

---

## 1. ACCESSIBILITY ISSUES (High Priority)

### 1.1 Missing ARIA Labels and Descriptions

**File:** `views/account.templ` (Lines 35-40)
- **Issue:** Checkbox input for "Enable daily reading notifications" has no explicit label connection
- **Fix:** Add `id` to checkbox and proper `for` attribute on label

**File:** `views/account.templ` (Lines 54-58)
- **Issue:** Time input field has no proper `id` attribute for label association
- **Fix:** Add `id="notification_time"` to input and `for="notification_time"` to label

**File:** `views/plans.templ` (Lines 13-19)
- **Issue:** Dropdown button missing `aria-label` for the "Create New Plan" button
- **Fix:** Add `aria-label="Create new reading plan"`

### 1.2 Non-Semantic Button Styling

**File:** `views/plans.templ` (Lines 14-19)
- **Issue:** Dropdown trigger uses `<div tabindex="0" role="button">` instead of `<button>` element
- **Fix:** Use native `<button>` element with appropriate classes

**File:** `views/notifications.templ` (Lines 8-16)
- **Issue:** Notification bell uses `<div tabindex="0" role="button">` instead of `<button>`
- **Fix:** Convert to `<button>` element for proper semantics

### 1.3 Incomplete ARIA Attributes

**File:** `views/plans.templ` (Lines 434-441)
- **Issue:** Button labeled "Select date" styled as input field has no explicit aria-label
- **Fix:** Add `aria-label="Select reading date"`

**File:** `views/layout.templ` (Line 28)
- **Issue:** Skip-to-content link should have more obvious focus styling
- **Fix:** Add `focus:ring` or `focus:outline` for better visibility

---

## 2. HTML STRUCTURE & SEMANTIC ISSUES (Medium Priority)

### 2.1 Form Control Pattern Violations

**File:** `views/plans.templ` (Lines 194-198)
- **Issue:** Uses `<label class="input input-bordered flex items-center gap-2">` wrapper around input, mixing label semantics with input styling
- **Fix:** Use separate label and input elements per CLAUDE.md guidelines

**File:** `views/auth.templ` (Lines 28-36)
- **Issue:** Password toggle button is inside a `<label>` element, violating HTML structure rules
- **Fix:** Restructure as separate form-control elements

### 2.2 Missing Form Associations

**File:** `views/account.templ` (Lines 54-59)
- **Issue:** `<input type="time" name="notification_time">` has no `id` attribute
- **Fix:** Add `id="notification_time"` and update label with `for` attribute

---

## 3. JAVASCRIPT & INTERACTIVITY ISSUES (Medium Priority)

### 3.1 Inline Event Handlers

**File:** `views/plans.templ` (Line 207)
- **Issue:** Uses inline `onchange` attribute with JavaScript
- **Fix:** Move to event listener in layout.templ or external JS file

**File:** `views/account.templ` (Lines 86, 95)
- **Issue:** Uses inline `onclick` attributes for push notification functions
- **Fix:** Use event listeners or HTMX attributes

**File:** `views/plans.templ` (Line 436)
- **Issue:** Button with inline `onclick="document.getElementById('date-picker-modal').showModal()"`
- **Fix:** Use templ script function pattern like `onclick={ showModal(...) }`

### 3.2 Code Duplication - Script Functions

**Files:** `views/plans.templ` (Lines 56-62), `views/auth.templ` (Lines 127-142)
- **Issue:** `showModal()`, `closeModal()`, `togglePassword()` defined in individual files
- **Fix:** Consider creating a shared utilities file or including in layout.templ

### 3.3 Improper Error Handling in JavaScript

**File:** `static/push-setup.js` (Lines 65, 73, 99, 103, 127)
- **Issue:** Multiple `alert()` calls for user-facing errors - poor UX
- **Fix:** Use toast notifications or in-page error messages using DaisyUI alert component

**File:** `static/push-setup.js` (Line 90-91)
- **Issue:** Uses `btoa()` and `String.fromCharCode.apply(null, ...)` which may fail with large arrays
- **Fix:** Use `Array.from()` method for better browser compatibility

**File:** `static/push-setup.js` (Lines 113-121)
- **Issue:** `disablePushNotifications()` doesn't check if unsubscribe succeeds before calling backend
- **Fix:** Add proper error handling and success validation

### 3.4 Dead Code

**File:** `views/layout.templ` (Lines 126-134)
- **Issue:** Event listener for "toggle-rename" class that handles toggling hidden elements - no elements with this class exist
- **Fix:** Remove dead code or document intended use

---

## 4. DAISYUI & TAILWINDCSS PATTERN ISSUES (Low Priority)

### 4.1 Inconsistent Visibility Toggle Patterns

**File:** `views/auth.templ` (Lines 31-34, 96-99)
- **Issue:** Uses hardcoded "hidden" class toggle for password visibility instead of proper pattern
- **Fix:** Consider using CSS transitions or aria-pressed state for better UX

**File:** `views/account.templ` (Lines 87, 96)
- **Issue:** Buttons use `hidden` class initially, revealed by JavaScript - no fallback if JS fails
- **Fix:** Use CSS `:disabled` state or HTMX to load elements conditionally

### 4.2 Improper Use of SVG Icons

**File:** `views/plans.templ` (Line 17-19)
- **Issue:** Dropdown toggle includes raw SVG instead of using icon component pattern
- **Fix:** Extract to `DropdownChevronIcon()` component

**File:** `views/plans.templ` (Lines 195-197)
- **Issue:** Another inline SVG in form input
- **Fix:** Replace with `@FileIcon()` templ component

### 4.3 Modal Dialog Pattern Inconsistency

**File:** `views/plans.templ` (Lines 127-148, 358-380)
- **Issue:** Modals mix `onclick={ showModal(...) }` templ pattern with plain `onclick="document.getElementById..."`
- **Fix:** Use consistent templ script function pattern for all modals

---

## 5. MISSING FEATURES & UX ISSUES (Medium Priority)

### 5.1 Incomplete Form Validation

**File:** `views/plans.templ` (Line 207)
- **Issue:** File input has `required` but no client-side validation feedback before submission
- **Fix:** Add file type validation and size checking before submission

### 5.2 Dead Links

**File:** `views/notifications.templ` (Lines 46, 57)
- **Issue:** Notification dropdown items and "View Dashboard" button link to `href="/"` instead of `/dashboard`
- **Fix:** Change both to `href="/dashboard"`

### 5.3 Confusing Button Styling

**File:** `views/plans.templ` (Lines 434-441)
- **Issue:** Button styled as text input field for date picker - misleading
- **Fix:** Use proper button styling or show visual indicator (e.g., `cursor-pointer`, `hover:bg-*`)

---

## 6. CSS & STYLING ISSUES (Low Priority)

### 6.1 Responsive Design Issues

**File:** `views/layout.templ` (Line 34)
- **Issue:** Mobile menu button uses `lg:hidden` but no visible focus indicator styling
- **Fix:** Ensure focus styling is visible with dark theme

### 6.2 Text Truncation Without Overflow Indicators

**File:** `views/notifications.templ` (Lines 49-50)
- **Issue:** Uses `truncate` class but no title attribute or tooltip to show full text
- **Fix:** Add `title` attribute with full text for accessibility

### 6.3 Missing Theme Contrast Validation

**File:** `input.css` (Lines 6-75)
- **Issue:** Custom DaisyUI theme "booky" uses oklch colors without WCAG contrast validation
- **Fix:** Verify all color combinations meet WCAG AA standards

### 6.4 Inconsistent Spacing

**File:** `views/plans.templ` (Line 422)
- **Issue:** Manual reading table cells have inconsistent spacing on mobile
- **Fix:** Test and fix responsive behavior on mobile devices

---

## 7. STATE & DATA MANAGEMENT ISSUES (Medium Priority)

### 7.1 Race Condition in CSV Import UI

**File:** `views/plans.templ` (Lines 70-109)
- **Issue:** Plan card shows processing state with `hx-get="/plans" hx-trigger="every 2s"` which reloads entire list
- **Fix:** Use targeted HTMX swap (`hx-target` should be specific plan card ID)

### 7.2 Form Reset on Successful Request

**File:** `views/plans.templ` (Line 499)
- **Issue:** Form resets on success but no loading/disabled state on button
- **Fix:** Add `hx-disabled-elt="find button"` to disable submit button during request

---

## 8. PERFORMANCE ISSUES (Medium Priority)

### 8.1 Unoptimized HTMX Polling

**File:** `views/plans.templ` (Line 106)
- **Issue:** Polling every 2 seconds during CSV import - generates many unnecessary requests
- **Fix:** Use exponential backoff, websocket, or increase polling interval to 5-10 seconds

### 8.2 Inefficient Icon Rendering

**File:** `views/icons.templ`
- **Issue:** All 21 SVG icons embedded directly - no lazy loading
- **Fix:** Consider using SVG sprite sheet for better caching

---

## 9. LAYOUT & UX STRUCTURE ISSUES (Low Priority)

### 9.1 Missing Breadcrumb Navigation

**File:** `views/plans.templ` (Lines 175-188, 245-254, 385-397)
- **Issue:** Create and Edit plan pages have back buttons but no clear breadcrumb
- **Fix:** Add breadcrumb or clear navigation indicators

### 9.2 Ambiguous Button Labels

**File:** `views/history.templ` (Line 50)
- **Issue:** Inconsistent action button labels - "Complete" vs "Completed"
- **Fix:** Standardize action button labels across all pages

### 9.3 Missing Loading States

**File:** `views/dashboard.templ` (Lines 69-74)
- **Issue:** Complete button shows loading spinner but no disabled state during request
- **Fix:** Add `hx-disabled-elt="this"` to disable button during request

---

## Priority Fix Order

**Immediate (Critical):**
1. Fix dead links in notifications.templ (lines 46, 57)
2. Remove dead code in layout.templ (toggle-rename listener)
3. Fix inline onclick consistency in plans.templ

**High Priority:**
1. Add ARIA labels to all interactive elements
2. Convert div[role="button"] to actual button elements
3. Add proper form label associations
4. Add loading/disabled states to all form submissions

**Medium Priority:**
1. Consolidate script functions into layout.templ
2. Replace alert() calls with DaisyUI toast notifications
3. Add text truncation titles for accessibility
4. Fix HTMX polling to be more efficient

**Low Priority:**
1. Extract inline SVGs to icon components
2. Add breadcrumb navigation
3. Standardize button labels
4. Validate theme color contrast
