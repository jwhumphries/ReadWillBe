package main

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"readwillbe/types"
)

func TestGetUserByID(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")
	plan := createTestPlan(t, db, user, "Test Plan")

	t.Run("existing user with plans", func(t *testing.T) {
		result, err := getUserByID(db, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email, result.Email)
		assert.Len(t, result.Plans, 1)
		assert.Equal(t, plan.Title, result.Plans[0].Title)
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, err := getUserByID(db, 9999)
		assert.Error(t, err)
	})
}

func TestUserExists(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "existing@example.com", "password123")

	t.Run("existing user", func(t *testing.T) {
		exists := userExists(user.Email, db)
		assert.True(t, exists)
	})

	t.Run("non-existent user", func(t *testing.T) {
		exists := userExists("nonexistent@example.com", db)
		assert.False(t, exists)
	})
}

func TestSignUpWithEmailAndPassword(t *testing.T) {
	db := setupTestDB(t)
	cfg := types.Config{AllowSignup: true}

	t.Run("successful signup", func(t *testing.T) {
		e := echo.New()
		form := url.Values{}
		form.Set("name", "John Doe")
		form.Set("email", "john@example.com")
		form.Set("password", "securepassword123")

		req := httptest.NewRequest("POST", "/auth/sign-up", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := signUpWithEmailAndPassword(db, cfg)
		err := handler(c)
		require.NoError(t, err)

		var user types.User
		err = db.First(&user, "email = ?", "john@example.com").Error
		require.NoError(t, err)
		assert.Equal(t, "John Doe", user.Name)

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("securepassword123"))
		assert.NoError(t, err)
	})

	t.Run("duplicate email", func(t *testing.T) {
		createTestUser(t, db, "duplicate@example.com", "password123")

		e := echo.New()
		form := url.Values{}
		form.Set("name", "Duplicate User")
		form.Set("email", "duplicate@example.com")
		form.Set("password", "anotherpassword123")

		req := httptest.NewRequest("POST", "/auth/sign-up", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := signUpWithEmailAndPassword(db, cfg)
		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 422, rec.Code)
	})

	t.Run("password too short", func(t *testing.T) {
		e := echo.New()
		form := url.Values{}
		form.Set("name", "Short Pass")
		form.Set("email", "short@example.com")
		form.Set("password", "short")

		req := httptest.NewRequest("POST", "/auth/sign-up", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := signUpWithEmailAndPassword(db, cfg)
		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 422, rec.Code)
	})

	t.Run("invalid email", func(t *testing.T) {
		e := echo.New()
		form := url.Values{}
		form.Set("name", "Invalid Email")
		form.Set("email", "not-an-email")
		form.Set("password", "validpassword123")

		req := httptest.NewRequest("POST", "/auth/sign-up", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := signUpWithEmailAndPassword(db, cfg)
		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 422, rec.Code)
	})

	t.Run("name too long", func(t *testing.T) {
		e := echo.New()
		form := url.Values{}
		form.Set("name", strings.Repeat("a", 300))
		form.Set("email", "toolong@example.com")
		form.Set("password", "validpassword123")

		req := httptest.NewRequest("POST", "/auth/sign-up", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := signUpWithEmailAndPassword(db, cfg)
		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 422, rec.Code)
	})
}

func TestSignInWithEmailAndPassword(t *testing.T) {
	db := setupTestDB(t)
	cfg := types.Config{}
	createTestUser(t, db, "login@example.com", "correctpassword123")

	t.Run("successful login", func(t *testing.T) {
		e := echo.New()
		form := url.Values{}
		form.Set("email", "login@example.com")
		form.Set("password", "correctpassword123")

		req := httptest.NewRequest("POST", "/auth/sign-in", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := signInWithEmailAndPassword(db, cfg)
		err := handler(c)
		require.NoError(t, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		e := echo.New()
		form := url.Values{}
		form.Set("email", "login@example.com")
		form.Set("password", "wrongpassword")

		req := httptest.NewRequest("POST", "/auth/sign-in", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := signInWithEmailAndPassword(db, cfg)
		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 422, rec.Code)
	})

	t.Run("non-existent user", func(t *testing.T) {
		e := echo.New()
		form := url.Values{}
		form.Set("email", "nonexistent@example.com")
		form.Set("password", "anypassword123")

		req := httptest.NewRequest("POST", "/auth/sign-in", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := signInWithEmailAndPassword(db, cfg)
		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 422, rec.Code)
	})

	t.Run("invalid email format", func(t *testing.T) {
		e := echo.New()
		form := url.Values{}
		form.Set("email", "not-an-email")
		form.Set("password", "anypassword123")

		req := httptest.NewRequest("POST", "/auth/sign-in", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := signInWithEmailAndPassword(db, cfg)
		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 422, rec.Code)
	})
}
