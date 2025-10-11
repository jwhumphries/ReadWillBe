package main

import (
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/gorilla/sessions"
	"readwillbe/types"
	"readwillbe/views"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func getUserByID(db *gorm.DB, id uint) (types.User, error) {
	var user types.User
	err := db.Preload("Plans").First(&user, "id = ?", id).Error

	return user, errors.Wrap(err, "Finding user")
}

func userExists(email string, db *gorm.DB) bool {
	var user types.User
	err := db.First(&user, "email = ?", email).Error

	return err != gorm.ErrRecordNotFound
}

func signUp() echo.HandlerFunc {
	return func(c echo.Context) error {
		return render(c, 200, views.SignUpForm(nil))
	}
}

func signUpWithEmailAndPassword(db *gorm.DB, cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")
		password := c.FormValue("password")

		parsedEmail, err := mail.ParseAddress(email)
		if err != nil {
			return render(c, 422, views.SignUpForm(fmt.Errorf("Invalid email address")))
		}
		email = parsedEmail.Address

		if userExists(email, db) {
			return render(c, 422, views.SignUpForm(fmt.Errorf("Email already registered")))
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return render(c, 422, views.SignUpForm(fmt.Errorf("Internal server error")))
		}

		user := types.User{
			Name:      name,
			Email:     email,
			Password:  string(hash),
			CreatedAt: time.Now(),
		}

		if err := db.Create(&user).Error; err != nil {
			err := errors.Wrap(err, "Create user error")
			return render(c, 422, views.SignUpForm(err))
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
			return render(c, 422, views.SignUpForm(errors.Wrap(err, "Internal server error")))
		}

		return c.Redirect(http.StatusFound, "/")
	}
}

func signIn(cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		return render(c, 200, views.SignInForm(cfg, nil))
	}
}

func signInWithEmailAndPassword(db *gorm.DB, cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		_, err := mail.ParseAddress(email)
		if err != nil {
			return render(c, 422, views.SignInForm(cfg, fmt.Errorf("Invalid email")))
		}

		var user types.User
		db.First(&user, "email = ?", email)
		if compareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); compareErr != nil {
			return render(c, 422, views.SignInForm(cfg, fmt.Errorf("Invalid email or password")))
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
			return render(c, 422, views.SignInForm(cfg, errors.Wrap(err, "Internal server error")))
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
