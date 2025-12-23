package main

import (
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"readwillbe/types"
	"readwillbe/views"
)

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
		return render(c, 200, views.SignUpPage(cfg, nil))
	}
}

func signUpWithEmailAndPassword(db *gorm.DB, cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")
		password := c.FormValue("password")

		parsedEmail, err := mail.ParseAddress(email)
		if err != nil {
			return render(c, 422, views.SignUpPage(cfg, fmt.Errorf("invalid email address")))
		}
		email = parsedEmail.Address

		if userExists(email, db) {
			return render(c, 422, views.SignUpPage(cfg, fmt.Errorf("email already registered")))
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return render(c, 422, views.SignUpPage(cfg, fmt.Errorf("internal server error")))
		}

		user := types.User{
			Name:      name,
			Email:     email,
			Password:  string(hash),
			CreatedAt: time.Now(),
		}

		if err := db.Create(&user).Error; err != nil {
			err := errors.Wrap(err, "Create user error")
			return render(c, 422, views.SignUpPage(cfg, err))
		}

		sess, _ := session.Get(SessionKey, c)
		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600 * 24 * 365,
			HttpOnly: true,
		}

		sess.Values[SessionUserIDKey] = user.ID

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return render(c, 422, views.SignUpPage(cfg, errors.Wrap(err, "Internal server error")))
		}

		return c.Redirect(http.StatusFound, "/")
	}
}

func signIn(cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		return render(c, 200, views.SignInPage(cfg, nil))
	}
}

func signInWithEmailAndPassword(db *gorm.DB, cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		_, err := mail.ParseAddress(email)
		if err != nil {
			return render(c, 422, views.SignInPage(cfg, fmt.Errorf("invalid email")))
		}

		var user types.User
		db.First(&user, "email = ?", email)
		if compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); compareErr != nil {
			return render(c, 422, views.SignInPage(cfg, fmt.Errorf("invalid email or password")))
		}

		sess, _ := session.Get(SessionKey, c)
		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600 * 24 * 365,
			HttpOnly: true,
		}

		sess.Values[SessionUserIDKey] = user.ID

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return render(c, 422, views.SignInPage(cfg, errors.Wrap(err, "Internal server error")))
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
