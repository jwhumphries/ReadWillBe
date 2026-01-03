# Backend Code Review - ReadWillBe

## Summary

| Severity | Count |
|----------|-------|
| Critical | 5 |
| High | 4 |
| Medium | 6 |
| Low/Quality | 10 |

---

## 1. CRITICAL ISSUES

### 1.1 Session Get Error Handling Ignored

**Files:**
- `cmd/readwillbe/plans.go` (lines 46, 56, 63)
- `cmd/readwillbe/user.go` (lines 72, 112, 132)
- `cmd/readwillbe/server.go` (line 199)

**Issue:** Session errors are silently discarded with blank identifiers:
```go
sess, _ := session.Get(SessionKey, c)
```

**Impact:** Session management failures could be hidden, leading to silent authentication issues.

**Fix:** Check and handle session errors appropriately.

---

### 1.2 Missing Error Check on db.First() in Sign-In Handler

**File:** `cmd/readwillbe/user.go` (line 107)

**Issue:** The db.First() call doesn't check for errors:
```go
var user types.User
db.First(&user, "email = ?", email)
if compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); compareErr != nil {
```

**Impact:** If user doesn't exist, comparison fails silently with empty password hash. Could lead to misleading error messages.

**Fix:**
```go
if err := db.First(&user, "email = ?", email).Error; err != nil {
    return render(c, 422, views.SignInPage(cfg, fmt.Errorf("invalid email or password")))
}
```

---

### 1.3 Type Assertion Errors Ignored

**File:** `cmd/readwillbe/plans.go` (lines 47-48)

**Issue:** Type assertions on session values ignore the success flag:
```go
title, _ := sess.Values[DraftTitleKey].(string)
readings, _ := sess.Values[DraftReadingsKey].([]views.ManualReading)
```

**Impact:** Could silently return empty/default values if stored data is corrupted or wrong type.

**Fix:** Check the boolean return value and handle type assertion failures.

---

### 1.4 Background Goroutine Database Connection Sharing

**File:** `cmd/readwillbe/plans.go` (lines 335-372)

**Issue:** Background CSV processing goroutine receives `*gorm.DB` by value and continues after HTTP context may be cancelled.

**Impact:**
- Orphaned database connections
- Operations continuing after user disconnects
- Potential data inconsistency if connection is closed

**Fix:** Pass context and use context-aware database operations.

---

### 1.5 Panic Recovery Only in Background Goroutine

**File:** `cmd/readwillbe/plans.go` (lines 336-343)

**Issue:** Panic recovery only catches errors within the CSV processing goroutine, not the main handler.

**Impact:** Unhandled panics could crash the application.

**Fix:** Add panic recovery middleware or handle panics in main handlers.

---

## 2. HIGH SEVERITY ISSUES

### 2.1 Session Cookie Missing Security Flags

**Files:**
- `cmd/readwillbe/user.go` (lines 73-77, 113-117)
- `cmd/readwillbe/server.go` (lines 223-227)

**Issue:** Session options missing `SameSite` attribute:
```go
sess.Options = &sessions.Options{
    Path:     "/",
    MaxAge:   3600 * 24 * 365,
    HttpOnly: true,
    // Missing: SameSite, Secure flags
}
```

**Impact:** Vulnerable to CSRF attacks; MaxAge of 1 year is excessive.

**Fix:**
```go
sess.Options = &sessions.Options{
    Path:     "/",
    MaxAge:   3600,  // 1 hour
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteStrictMode,
}
```

---

### 2.2 Notification Worker Goroutine Never Terminates

**File:** `cmd/readwillbe/push.go` (lines 96-134)

**Issue:** Notification worker goroutine runs indefinitely with no shutdown mechanism:
```go
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        // ... notification logic
    }
}()
```

**Impact:** Prevents proper server termination during graceful shutdown.

**Fix:** Add context-based cancellation or shutdown channel.

---

### 2.3 Query N+1 Problem in Handlers

**Files:**
- `cmd/readwillbe/notifications.go` (lines 20-30)
- `cmd/readwillbe/dashboard.go` (lines 20-22)
- `cmd/readwillbe/push.go` (lines 117-119)

**Issue:** Inefficient querying pattern loads all readings then filters in memory:
```go
db.Where("plan_id IN (?)",
    db.Table("plans").Select("id").Where("user_id = ?", user.ID),
).Find(&readings)
// Then filters by IsActiveToday() in Go code
```

**Impact:** For large datasets, loads unnecessary data into memory.

**Fix:** Use SQL LIMIT/ORDER BY clauses to filter at database level.

---

### 2.4 Deprecated Error Pattern Still Present

**File:** `types/config.go` (lines 44-46)

**Issue:** Function marked as deprecated but still importable:
```go
func ConfigFromEnv() (Config, error) {
    return Config{}, fmt.Errorf("ConfigFromEnv is deprecated, use ConfigFromViper instead")
}
```

**Impact:** Could be accidentally used.

**Fix:** Remove entirely or use build constraints.

---

## 3. MEDIUM SEVERITY ISSUES

### 3.1 Inconsistent Error Handling in Database Operations

**File:** `cmd/readwillbe/user.go` (lines 19-24)

**Issue:** `getUserByID()` wraps all errors indiscriminately:
```go
func getUserByID(db *gorm.DB, id uint) (types.User, error) {
    var user types.User
    err := db.Preload("Plans").First(&user, "id = ?", id).Error
    return user, errors.Wrap(err, "Finding user")
}
```

**Impact:** Loses distinction between "user not found" and actual database errors.

**Fix:** Handle `gorm.ErrRecordNotFound` separately from other errors.

---

### 3.2 Missing Error Check in File Close

**File:** `cmd/readwillbe/plans.go` (line 94)

**Issue:** File close error is ignored:
```go
f.Close() // Close immediately after reading
```

**Fix:** Use `defer` with error handling or check error explicitly.

---

### 3.3 Inconsistent Error Logging

**Files:**
- `cmd/readwillbe/push.go` (uses fmt.Printf/Println)
- `cmd/readwillbe/seed_dev.go` (uses logrus)
- `cmd/readwillbe/config.go` (uses fmt.Println())

**Issue:** Mixing `fmt.Printf()` and `logrus` for error logging.

**Impact:** Makes log aggregation difficult and inconsistent.

**Fix:** Use logrus consistently for all error logging.

---

### 3.4 Missing Error Handling in Test Helpers

**File:** `cmd/readwillbe/plans_test.go` (lines 67-68)

**Issue:** Test ignores form file creation errors:
```go
part, _ := writer.CreateFormFile("csv", "readings.csv")
_, _ = part.Write([]byte(csvContent))
```

**Fix:** Check errors in tests using `require.NoError()`.

---

### 3.5 Render Error Handling Could Be Improved

**File:** `cmd/readwillbe/server.go` (lines 28-36)

**Issue:** Template rendering errors return generic message without logging:
```go
if err != nil {
    return ctx.String(http.StatusInternalServerError, "failed to render response template")
}
```

**Impact:** Debugging difficult without error context.

**Fix:** Log the error before returning.

---

### 3.6 No Input Validation on Time Format

**File:** `cmd/readwillbe/account.go` (line 31)

**Issue:** Notification time accepted as string without validation:
```go
notificationTime := c.FormValue("notification_time")
```

**Impact:** Invalid time formats could cause issues later.

**Fix:** Validate "HH:MM" format before saving.

---

## 4. LOW SEVERITY / CODE QUALITY ISSUES

### 4.1 Hardcoded Bcrypt Cost Factor

**Files:**
- `cmd/readwillbe/user.go` (line 55)
- `cmd/readwillbe/seed_dev.go` (line 34)

**Issue:** Bcrypt cost is hardcoded:
```go
hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
```

**Fix:** Make it a constant or configuration parameter.

---

### 4.2 Repeated Authentication Pattern

**Files:** Multiple handler files

**Issue:** Many handlers repeat:
```go
user, ok := GetSessionUser(c)
if !ok {
    return c.Redirect(http.StatusFound, "/auth/sign-in")
}
```

**Fix:** Create middleware that automatically redirects unauthenticated requests.

---

### 4.3 Missing Preload in Some Queries

**File:** `cmd/readwillbe/notifications.go` (line 20)

**Issue:** `notificationCount` doesn't preload Plan while `notificationDropdown` does (line 44).

**Fix:** Ensure consistent preloading strategy.

---

### 4.4 Magic Numbers and String Constants

**File:** `cmd/readwillbe/push.go` (line 97)

**Issue:** Hardcoded ticker duration:
```go
ticker := time.NewTicker(1 * time.Minute)
```

**Fix:** Extract to constant for easier configuration/testing.

---

### 4.5 Session Options Duplication

**Files:**
- `cmd/readwillbe/user.go` (lines 73-77, 113-117)
- `cmd/readwillbe/server.go` (lines 223-227)

**Issue:** Session options set identically in three places.

**Fix:** Extract to helper function or constant.

---

### 4.6 Race Condition in Notification Worker

**File:** `cmd/readwillbe/push.go` (lines 117-119)

**Issue:** Database query in goroutine without proper error handling.

**Fix:** Add error handling for database operations in background goroutine.

---

### 4.7 Missing Content-Type Header Handling

**File:** `cmd/readwillbe/csv.go`

**Issue:** CSV parsing doesn't validate Content-Type, potentially accepting non-CSV files.

**Fix:** Verify actual file content matches expected type.

---

### 4.8 Potential Memory Issue in Notifications

**File:** `cmd/readwillbe/push.go` (lines 105-120)

**Issue:** Each minute, worker queries all users with notifications, preloads subscriptions, and loads all readings.

**Impact:** With many users, could accumulate memory without cleanup.

**Fix:** Implement pagination or batch processing.

---

### 4.9 Context Usage in Goroutines

**File:** `cmd/readwillbe/plans.go` (line 335)

**Issue:** Goroutine receives database connection but no context for cancellation.

**Fix:** Pass and use context for graceful cancellation.

---

### 4.10 Test Isolation Could Be Improved

**File:** `cmd/readwillbe/plans_test.go`

**Issue:** Tests use real database with memory-based SQLite but isolation could be better.

**Fix:** Consider using fixtures or factories for repeated setup.

---

## 5. GOOD PATTERNS OBSERVED

- **Proper GORM transactions** - Multiple places correctly use `db.Transaction()`
- **SQL parameterization** - All queries use parameterized queries with `?` placeholders
- **Proper middleware chain** - Good use of Echo middleware for compression, logging, security
- **Password hashing** - Using bcrypt correctly
- **Database connection pooling** - Properly configured with pool settings
- **Error wrapping** - Good use of `pkg/errors` for error context
- **User caching** - Implements TTL-based user cache to reduce database load
- **Comprehensive test coverage** - CSV parsing and plan creation have tests

---

## 6. OPTIMIZATION OPPORTUNITIES

1. **Query optimization** - Implement SQL-level LIMIT/OFFSET for notification queries
2. **Caching strategy** - Consider caching plan counts to avoid repeated database queries
3. **Async notification sending** - Send notifications asynchronously in queue instead of goroutine
4. **Connection reuse** - Background goroutine should use connection pool properly
5. **Reduce redundant queries** - Some endpoints query same data multiple times

---

## Priority Fix Order

**Immediate (Critical):**
1. Fix missing error check in sign-in handler (user.go:107)
2. Add session error handling (all session.Get calls)
3. Fix type assertion error handling (plans.go:47-48)
4. Add graceful shutdown to notification worker (push.go:96)

**High Priority:**
1. Add SameSite and Secure flags to session cookies
2. Fix N+1 query issues in notifications and dashboard
3. Add context-based cancellation to background goroutines

**Medium Priority:**
1. Standardize error logging to use logrus
2. Add input validation for time format
3. Extract session options to shared configuration
4. Add proper error handling in render function

**Low Priority:**
1. Extract bcrypt cost to constant
2. Create authentication middleware
3. Remove deprecated ConfigFromEnv function
4. Add Content-Type validation for CSV uploads
