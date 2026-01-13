package main

import (
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"readwillbe/types"
	"readwillbe/views"
)

const (
	MaxNameLength     = 255
	MaxEmailLength    = 254
	MinPasswordLength = 12
	MaxPasswordLength = 128
	BcryptCost        = 12
)

// dummyHash is a valid bcrypt hash used for timing attack prevention.
// Generated with cost 12 to match BcryptCost for consistent timing.
var dummyHash = []byte("$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.qfTJ.l5NmY8S2e")

func getUserByID(db *gorm.DB, id uint) (types.User, error) {
	var user types.User
	err := db.Preload("Plans").First(&user, "id = ?", id).Error

	return user, errors.Wrap(err, "Finding user")
}

func userExists(email string, db *gorm.DB) bool {
	var user types.User
	err := db.First(&user, "email = ?", email).Error

	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func signUp(cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		csrf, _ := c.Get("csrf").(string)
		return render(c, 200, views.SignUpPage(cfg, csrf, nil))
	}
}

func signUpWithEmailAndPassword(db *gorm.DB, cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")
		password := c.FormValue("password")
		csrf, _ := c.Get("csrf").(string)

		if len(name) == 0 || len(name) > MaxNameLength {
			return render(c, 422, views.SignUpPage(cfg, csrf, fmt.Errorf("name must be between 1 and %d characters", MaxNameLength)))
		}

		if len(email) > MaxEmailLength {
			return render(c, 422, views.SignUpPage(cfg, csrf, fmt.Errorf("email must be less than %d characters", MaxEmailLength)))
		}

		parsedEmail, err := mail.ParseAddress(email)
		if err != nil {
			return render(c, 422, views.SignUpPage(cfg, csrf, fmt.Errorf("invalid email address")))
		}
		email = parsedEmail.Address

		if len(password) < MinPasswordLength || len(password) > MaxPasswordLength {
			return render(c, 422, views.SignUpPage(cfg, csrf, fmt.Errorf("password must be between %d and %d characters", MinPasswordLength, MaxPasswordLength)))
		}

		if userExists(email, db) {
			return render(c, 422, views.SignUpPage(cfg, csrf, fmt.Errorf("email already registered")))
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
		if err != nil {
			return render(c, 422, views.SignUpPage(cfg, csrf, fmt.Errorf("internal server error")))
		}

		user := types.User{
			Name:      name,
			Email:     email,
			Password:  string(hash),
			CreatedAt: time.Now(),
		}

		if dbErr := db.Create(&user).Error; dbErr != nil {
			wrappedErr := errors.Wrap(dbErr, "Create user error")
			return render(c, 422, views.SignUpPage(cfg, csrf, wrappedErr))
		}

		sess, err := session.Get(SessionKey, c)
		if err != nil {
			return render(c, 422, views.SignUpPage(cfg, csrf, fmt.Errorf("internal server error")))
		}
		sess.Options = getSecureSessionOptions(cfg)

		sess.Values[SessionUserIDKey] = user.ID

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return render(c, 422, views.SignUpPage(cfg, csrf, errors.Wrap(err, "Internal server error")))
		}

		return c.Redirect(http.StatusFound, "/")
	}
}

func signIn(cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		csrf, _ := c.Get("csrf").(string)
		return render(c, 200, views.SignInPage(cfg, csrf, nil))
	}
}

func signInWithEmailAndPassword(db *gorm.DB, cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		csrf, _ := c.Get("csrf").(string)

		_, err := mail.ParseAddress(email)
		if err != nil {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return render(c, 422, views.SignInPage(cfg, csrf, fmt.Errorf("invalid email or password")))
		}

		var user types.User
		if dbErr := db.First(&user, "email = ?", email).Error; dbErr != nil {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return render(c, 422, views.SignInPage(cfg, csrf, fmt.Errorf("invalid email or password")))
		}

		if user.Password == "" {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return render(c, 422, views.SignInPage(cfg, csrf, fmt.Errorf("invalid email or password")))
		}

		if compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); compareErr != nil {
			return render(c, 422, views.SignInPage(cfg, csrf, fmt.Errorf("invalid email or password")))
		}

		// Invalidate any existing session to prevent session fixation
		oldSess, _ := session.Get(SessionKey, c)
		if oldSess != nil {
			oldSess.Options.MaxAge = -1
			_ = oldSess.Save(c.Request(), c.Response())
		}

		// Create new session
		sess, err := session.Get(SessionKey, c)
		if err != nil {
			return render(c, 422, views.SignInPage(cfg, csrf, fmt.Errorf("internal server error")))
		}
		sess.Options = getSecureSessionOptions(cfg)

		sess.Values[SessionUserIDKey] = user.ID

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return render(c, 422, views.SignInPage(cfg, csrf, errors.Wrap(err, "Internal server error")))
		}

		return c.Redirect(http.StatusFound, "/")
	}
}

func signOut() echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get("session", c)
		sess.Options.MaxAge = -1
		err := sess.Save(c.Request(), c.Response())
		if err != nil {
			return err
		}

		return c.Redirect(http.StatusFound, "/")
	}
}
