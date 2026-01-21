package main

import (
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v5"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	mw "readwillbe/internal/middleware"
	"readwillbe/internal/model"
	"readwillbe/internal/repository"
	"readwillbe/internal/views"
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

func signUp(cfg model.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
		csrf, _ := c.Get("csrf").(string)
		return render(c, 200, views.SignUpPage(cfg, csrf, nil))
	}
}

func signUpWithEmailAndPassword(db *gorm.DB, cfg model.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
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

		if repository.UserExists(email, db) {
			return render(c, 422, views.SignUpPage(cfg, csrf, fmt.Errorf("email already registered")))
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
		if err != nil {
			return render(c, 422, views.SignUpPage(cfg, csrf, fmt.Errorf("internal server error")))
		}

		user := model.User{
			Name:      name,
			Email:     email,
			Password:  string(hash),
			CreatedAt: time.Now(),
		}

		if dbErr := repository.CreateUser(db, &user); dbErr != nil {
			wrappedErr := errors.Wrap(dbErr, "Create user error")
			return render(c, 422, views.SignUpPage(cfg, csrf, wrappedErr))
		}

		sess, err := session.Get(mw.SessionKey, c)
		if err != nil {
			return render(c, 422, views.SignUpPage(cfg, csrf, fmt.Errorf("internal server error")))
		}
		sess.Options = mw.GetSecureSessionOptions(cfg)

		sess.Values[mw.SessionUserIDKey] = user.ID

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return render(c, 422, views.SignUpPage(cfg, csrf, errors.Wrap(err, "Internal server error")))
		}

		return c.Redirect(http.StatusFound, "/")
	}
}

func signIn(cfg model.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
		csrf, _ := c.Get("csrf").(string)
		return render(c, 200, views.SignInPage(cfg, csrf, nil))
	}
}

func signInWithEmailAndPassword(db *gorm.DB, cfg model.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		csrf, _ := c.Get("csrf").(string)

		_, err := mail.ParseAddress(email)
		if err != nil {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return render(c, 422, views.SignInPage(cfg, csrf, fmt.Errorf("invalid email or password")))
		}

		user, dbErr := repository.GetUserByEmail(db, email)
		if dbErr != nil {
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

		// Clear existing session values to prevent session fixation
		sess, err := session.Get(mw.SessionKey, c)
		if err != nil {
			return render(c, 422, views.SignInPage(cfg, csrf, fmt.Errorf("internal server error")))
		}
		for key := range sess.Values {
			delete(sess.Values, key)
		}
		sess.Options = mw.GetSecureSessionOptions(cfg)
		sess.Values[mw.SessionUserIDKey] = user.ID

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return render(c, 422, views.SignInPage(cfg, csrf, errors.Wrap(err, "Internal server error")))
		}

		return c.Redirect(http.StatusFound, "/")
	}
}

func signOut() echo.HandlerFunc {
	return func(c *echo.Context) error {
		sess, _ := session.Get("session", c)
		sess.Options.MaxAge = -1
		err := sess.Save(c.Request(), c.Response())
		if err != nil {
			return err
		}

		return c.Redirect(http.StatusFound, "/")
	}
}
