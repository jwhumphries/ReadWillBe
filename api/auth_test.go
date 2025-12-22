//go:build !wasm

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"readwillbe/types"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	if err := db.AutoMigrate(&types.User{}, &types.Plan{}, &types.Reading{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}
	return db
}

func createTestUser(t *testing.T, db *gorm.DB, email, password string) types.User {
	t.Helper()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := types.User{
		Name:     "Test User",
		Email:    email,
		Password: string(hashedPassword),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestSignIn_Success(t *testing.T) {
	db := setupTestDB(t)
	cfg := types.Config{CookieSecret: []byte("test-secret-key-32-chars-long!!")}
	createTestUser(t, db, "test@example.com", "password123")

	e := echo.New()
	reqBody, _ := json.Marshal(SignInRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-in", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SignIn(db, cfg)
	if err := handler(c); err != nil {
		t.Errorf("SignIn failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if response.Token == "" {
		t.Error("expected token in response")
	}
	if response.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", response.User.Email)
	}
	if response.User.Password != "" {
		t.Error("password should be empty in response")
	}
}

func TestSignIn_InvalidCredentials(t *testing.T) {
	db := setupTestDB(t)
	cfg := types.Config{CookieSecret: []byte("test-secret-key-32-chars-long!!")}
	createTestUser(t, db, "test@example.com", "password123")

	e := echo.New()
	reqBody, _ := json.Marshal(SignInRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-in", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SignIn(db, cfg)
	err := handler(c)
	if err == nil {
		t.Error("expected error for invalid credentials")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, httpErr.Code)
	}
}

func TestSignIn_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	cfg := types.Config{CookieSecret: []byte("test-secret-key-32-chars-long!!")}

	e := echo.New()
	reqBody, _ := json.Marshal(SignInRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-in", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SignIn(db, cfg)
	err := handler(c)
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, httpErr.Code)
	}
}

func TestSignUp_Success(t *testing.T) {
	db := setupTestDB(t)
	cfg := types.Config{
		CookieSecret: []byte("test-secret-key-32-chars-long!!"),
		AllowSignup:  true,
	}

	e := echo.New()
	reqBody, _ := json.Marshal(SignUpRequest{
		Name:     "New User",
		Email:    "newuser@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SignUp(db, cfg)
	if err := handler(c); err != nil {
		t.Errorf("SignUp failed: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var response AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if response.Token == "" {
		t.Error("expected token in response")
	}
	if response.User.Email != "newuser@example.com" {
		t.Errorf("expected email newuser@example.com, got %s", response.User.Email)
	}
}

func TestSignUp_Disabled(t *testing.T) {
	db := setupTestDB(t)
	cfg := types.Config{
		CookieSecret: []byte("test-secret-key-32-chars-long!!"),
		AllowSignup:  false,
	}

	e := echo.New()
	reqBody, _ := json.Marshal(SignUpRequest{
		Name:     "New User",
		Email:    "newuser@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SignUp(db, cfg)
	err := handler(c)
	if err == nil {
		t.Error("expected error when signup is disabled")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, httpErr.Code)
	}
}

func TestSignUp_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	cfg := types.Config{
		CookieSecret: []byte("test-secret-key-32-chars-long!!"),
		AllowSignup:  true,
	}
	createTestUser(t, db, "existing@example.com", "password123")

	e := echo.New()
	reqBody, _ := json.Marshal(SignUpRequest{
		Name:     "New User",
		Email:    "existing@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/sign-up", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SignUp(db, cfg)
	err := handler(c)
	if err == nil {
		t.Error("expected error for duplicate email")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, httpErr.Code)
	}
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	db := setupTestDB(t)
	cfg := types.Config{CookieSecret: []byte("test-secret-key-32-chars-long!!")}
	user := createTestUser(t, db, "test@example.com", "password123")

	token, _, err := generateToken(user, cfg.CookieSecret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := JWTMiddleware(cfg)
	handler := middleware(func(c echo.Context) error {
		userID := c.Get("user_id").(uint)
		if userID != user.ID {
			t.Errorf("expected user_id %d, got %d", user.ID, userID)
		}
		return c.String(http.StatusOK, "ok")
	})

	if err := handler(c); err != nil {
		t.Errorf("middleware failed: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	cfg := types.Config{CookieSecret: []byte("test-secret-key-32-chars-long!!")}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := JWTMiddleware(cfg)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err == nil {
		t.Error("expected error for missing token")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, httpErr.Code)
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	cfg := types.Config{CookieSecret: []byte("test-secret-key-32-chars-long!!")}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := JWTMiddleware(cfg)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	if err == nil {
		t.Error("expected error for invalid token")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, httpErr.Code)
	}
}
