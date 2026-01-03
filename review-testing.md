# Testing Review - ReadWillBe

## Summary

| Metric | Value |
|--------|-------|
| Total Source Files | 25 |
| Test Files | 3 (12%) |
| Test Functions | 8 |
| Tested Source Files | 3 (12%) |
| Untested Files | 21 (84%) |
| Critical Business Logic Coverage | ~20% |
| Handler Coverage | <5% (1 of 30+ handlers) |

---

## 1. EXISTING TEST FILES

### 1.1 `cmd/readwillbe/csv_test.go` (165 lines)

**Tests:**
- `TestParseCSV()` - 5 test cases covering valid/invalid CSV parsing
- `TestParseDate()` - 11 test cases covering date format parsing

**Quality:** Good
- Well-structured table-driven tests
- Good error case coverage (4/5 cases)
- Tests edge cases (week 52, multiple date formats)

---

### 1.2 `cmd/readwillbe/plans_test.go` (168 lines)

**Tests:**
- `setupTestDB()` - Test database setup helper
- `TestCreatePlan_BackgroundProcessing()` - Integration test for async plan creation
- `TestCreatePlan_BackgroundProcessingFailure()` - Failure case test
- `TestUserCache()` - Cache functionality test

**Quality:** Good
- Proper use of `assert.Eventually()` for async handling
- Tests both success and failure paths
- Good test database setup with cleanup

---

### 1.3 `types/reading_test.go` (126 lines)

**Tests:**
- `TestReading_IsOverdue()` - 5 test cases for overdue logic
- `TestReading_IsActiveToday()` - 4 test cases for active today logic

**Quality:** Good
- Tests critical business logic methods
- Good use of relative dates
- Both methods have adequate coverage

---

## 2. UNTESTED FILES (21 of 25 = 84%)

### 2.1 Critical - Authentication System

**File:** `cmd/readwillbe/user.go` (142 lines)

**Untested Functions:**
- `getUserByID()` - DB query helper (line 19)
- `userExists()` - Existence check (line 26)
- `signUp()` - Signup form display (line 33)
- `signUpWithEmailAndPassword()` - Registration logic (line 39)
- `signIn()` - Login form display (line 90)
- `signInWithEmailAndPassword()` - Authentication (line 96)
- `signOut()` - Logout (line 130)

**Missing Test Coverage:**
- Password hashing verification
- Session creation/destruction
- Email validation
- Duplicate email prevention

---

### 2.2 Critical - Reading Status Management

**File:** `cmd/readwillbe/readings.go` (113 lines)

**Untested Functions:**
- `completeReading()` - Marks reading completed (line 13)
- `uncompleteReading()` - Reverts completion (line 46)
- `updateReading()` - Updates reading content (line 78)

**Missing Test Coverage:**
- Complete/uncomplete status changes
- CompletedAt timestamp validation
- Update with empty content validation
- Authorization checks

---

### 2.3 Critical - Push Notification System

**File:** `cmd/readwillbe/push.go` (186 lines)

**Untested Functions:**
- `saveSubscription()` - Saves push endpoint (line 23)
- `removeSubscription()` - Removes subscription (line 51)
- `removeAllSubscriptions()` - Removes all (line 74)
- `startNotificationWorker()` - Async worker loop (line 90)
- `sendNotification()` - Sends push notifications (line 139)

**Missing Test Coverage:**
- Subscription persistence
- Goroutine-based notification worker
- External API calls (webpush)
- Stale subscription cleanup (410 status)

---

### 2.4 Critical - Plan Management (Partial Coverage)

**File:** `cmd/readwillbe/plans.go` (556 lines)

**Tested:** Background CSV processing only (~140 lines)
**Untested:** ~416 lines (75%)

**Untested Functions:**
- `plansListHandler()` - List plans (line 17)
- `createPlanForm()` - Display form (line 31)
- `manualPlanForm()` - Manual creation form (line 69)
- `updateDraftTitle()` - Update draft title (line 81)
- `addDraftReading()` - Add reading to draft (line 97)
- `getDraftReading()` - Fetch draft reading (line 127)
- `getDraftReadingEdit()` - Display edit form (line 146)
- `updateDraftReading()` - Update draft reading (line 165)
- `deleteDraftReading()` - Delete from draft (line 199)
- `deleteDraft()` - Clear all draft data (line 224)
- `createManualPlan()` - Create plan from manual entry (line 238) - USES TRANSACTION
- `renamePlan()` - Rename existing plan (line 378)
- `deletePlan()` - Delete plan (line 409)
- `editPlanForm()` - Edit form display (line 434)
- `editPlan()` - Update plan data (line 455) - USES TRANSACTION
- `deleteReading()` - Remove reading (line 522)

---

### 2.5 High - Dashboard & History

**File:** `cmd/readwillbe/dashboard.go` (42 lines)
- `dashboardHandler()` - Fetches and filters today/overdue readings (line 12)

**File:** `cmd/readwillbe/history.go` (31 lines)
- `historyHandler()` - Fetches completed readings (line 12)

---

### 2.6 High - Notifications

**File:** `cmd/readwillbe/notifications.go` (66 lines)
- `notificationCount()` - Counts today's readings (line 12)
- `notificationDropdown()` - Fetches top 5 readings (line 36)

---

### 2.7 Medium - Account Settings

**File:** `cmd/readwillbe/account.go` (43 lines)
- `accountHandler()` - Renders account page (line 12)
- `updateSettings()` - Updates notification settings (line 23)

---

### 2.8 Low - Configuration

**File:** `types/config.go` (46 lines)
- `ConfigFromViper()` - Load config from viper (line 21)

**File:** `types/user.go` (26 lines)
- `User.IsSet()` - Check if user is valid (line 23)

---

### 2.9 Not Required - Infrastructure

- `main.go` - CLI entry point
- `root.go` - CLI root command
- `config.go` (cmd) - CLI config handling
- `server.go` - Server initialization
- `version.go` - Version display
- `seed_dev.go` / `seed_prod.go` - Data seeding

---

## 3. MISSING TEST PATTERNS

### 3.1 No Middleware Tests
- Session authentication checks
- User context extraction
- Authorization/permission validation

### 3.2 No Error Handling Tests
- Most error paths return strings via `c.String()`
- No validation of error messages
- No tests for DB connection failures

### 3.3 No Concurrent/Race Condition Tests
- Goroutine-based notification worker
- Cache operations (sync.Map)
- No goroutine leak tests

### 3.4 No Authorization Tests
- Most handlers check `GetSessionUser()` and redirect
- No tests for unauthorized access attempts
- No tests for accessing other users' plans/readings

### 3.5 No Database Transaction Tests
- `createManualPlan()` uses transaction (line 275-289)
- `editPlan()` uses transaction (line 484-512)
- No tests for rollback scenarios

### 3.6 No Mock/Stub Tests
- External webpush API calls not mocked
- Real database used in all tests
- No interface-based dependencies

---

## 4. TEST QUALITY ASSESSMENT

### Strengths
- Good test structure with table-driven tests
- Proper test database setup with cleanup
- Appropriate use of testify assertions
- Clear test names describing scenarios
- Async handling with `assert.Eventually()`
- Real integration tests using SQLite

### Weaknesses
- Massive coverage gaps (76% of files untested)
- Only 3 test files total
- No negative/authorization test cases
- Insufficient test isolation (no mocks)
- Missing failure scenarios
- No performance/load tests

---

## 5. RECOMMENDED TESTS TO ADD

### Priority 1: Critical Business Logic

**Authentication System (user.go):**
```go
func TestSignUpWithEmailAndPassword(t *testing.T)
func TestSignUpWithDuplicateEmail(t *testing.T)
func TestSignInWithValidCredentials(t *testing.T)
func TestSignInWithInvalidPassword(t *testing.T)
func TestSignInWithNonExistentUser(t *testing.T)
func TestSignOut(t *testing.T)
func TestPasswordHashing(t *testing.T)
```

**Reading Status Management (readings.go):**
```go
func TestCompleteReading(t *testing.T)
func TestCompleteReadingUpdatesTimestamp(t *testing.T)
func TestUncompleteReading(t *testing.T)
func TestUpdateReadingContent(t *testing.T)
func TestCompleteReadingAuthorization(t *testing.T)
```

**Plan Operations (plans.go):**
```go
func TestCreateManualPlan(t *testing.T)
func TestCreateManualPlanTransaction(t *testing.T)
func TestEditPlan(t *testing.T)
func TestDeletePlan(t *testing.T)
func TestRenamePlan(t *testing.T)
func TestPlanAuthorizationCrossUser(t *testing.T)
```

### Priority 2: Important Features

**Push Notifications (push.go):**
```go
func TestSaveSubscription(t *testing.T)
func TestRemoveSubscription(t *testing.T)
func TestNotificationWorkerStartup(t *testing.T)
func TestSendNotification_Success(t *testing.T)
func TestSendNotification_StaleSubscription(t *testing.T)
```

**Notification Count (notifications.go):**
```go
func TestNotificationCount(t *testing.T)
func TestNotificationDropdown(t *testing.T)
func TestNotificationDropdownLimit(t *testing.T)
```

### Priority 3: Authorization/Permissions

```go
func TestPlanAccessByOtherUser(t *testing.T)
func TestReadingAccessByOtherUser(t *testing.T)
func TestAccountAccessByOtherUser(t *testing.T)
func TestUnauthenticatedAccess(t *testing.T)
```

### Priority 4: Error Handling

```go
func TestDatabaseConnectionFailure(t *testing.T)
func TestInvalidFormData(t *testing.T)
func TestMalformedCSV(t *testing.T)
func TestEmptyFormSubmission(t *testing.T)
```

### Priority 5: Transaction Rollback

```go
func TestCreateManualPlanRollback(t *testing.T)
func TestEditPlanRollback(t *testing.T)
```

---

## 6. TEST INFRASTRUCTURE IMPROVEMENTS

### 6.1 Add Test Fixtures

Create `testdata/` directory with:
- Sample CSV files
- Expected reading outputs
- User fixtures

### 6.2 Add Test Helpers

```go
// In test_helpers.go
func createTestUser(t *testing.T, db *gorm.DB) *types.User
func createTestPlan(t *testing.T, db *gorm.DB, user *types.User) *types.Plan
func createTestReading(t *testing.T, db *gorm.DB, plan *types.Plan) *types.Reading
func authenticatedRequest(t *testing.T, user *types.User) echo.Context
```

### 6.3 Add Mock for External Services

```go
type MockWebPushClient interface {
    SendNotification(subscription webpush.Subscription, message []byte) error
}

type mockWebPush struct {
    sentNotifications [][]byte
    shouldFail        bool
}
```

### 6.4 Add Integration Test Suite

```go
func TestFullUserWorkflow(t *testing.T) {
    // Sign up -> Create plan -> Complete readings -> View history
}

func TestCSVImportToCompletion(t *testing.T) {
    // Upload CSV -> Wait for processing -> Verify readings -> Complete all
}
```

---

## 7. COVERAGE TARGET RECOMMENDATIONS

| Area | Current | Target |
|------|---------|--------|
| Authentication | 0% | 90% |
| Plans (handlers) | 25% | 80% |
| Readings | 0% | 85% |
| Push Notifications | 0% | 70% |
| Dashboard/History | 0% | 60% |
| Notifications | 0% | 70% |
| Account | 0% | 60% |
| **Overall** | **~15%** | **75%** |

---

## 8. PRIORITY ORDER

**Immediate:**
1. Add tests for authentication system (user.go)
2. Add tests for reading status management (readings.go)
3. Add tests for plan operations (plans.go)

**High:**
4. Add tests for notification system (push.go, notifications.go)
5. Add authorization/permission tests
6. Add error handling tests

**Medium:**
7. Add transaction rollback tests
8. Add concurrent access tests for cache
9. Add configuration/startup tests

**Low:**
10. Add performance/benchmark tests
11. Add integration test suite
12. Mock external services (WebPush API)
