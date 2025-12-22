//go:build !wasm

package api

import (
	"net/http"
	"strings"
	"time"

	"readwillbe/types"

	"github.com/charmbracelet/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type AuthResponse struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	User      types.User `json:"user"`
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func generateToken(user types.User, secret []byte) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "signing token")
	}
	return tokenString, expiresAt, nil
}

func JWTMiddleware(cfg types.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Debug("missing authorization header")
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return cfg.CookieSecret, nil
			})

			if err != nil {
				log.Debug("failed to parse token", "error", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			if !token.Valid {
				log.Debug("token is not valid")
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			return next(c)
		}
	}
}

func SignIn(db *gorm.DB, cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req SignInRequest
		if err := c.Bind(&req); err != nil {
			log.Debug("failed to bind sign in request", "error", err)
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		var user types.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			log.Debug("user not found", "email", req.Email)
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			log.Debug("password mismatch", "email", req.Email)
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
		}

		token, expiresAt, err := generateToken(user, cfg.CookieSecret)
		if err != nil {
			log.Error("failed to generate token", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
		}

		user.Password = ""
		log.Info("user signed in", "email", user.Email)
		return c.JSON(http.StatusOK, AuthResponse{
			Token:     token,
			ExpiresAt: expiresAt,
			User:      user,
		})
	}
}

func SignUp(db *gorm.DB, cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !cfg.AllowSignup {
			return echo.NewHTTPError(http.StatusForbidden, "signup is disabled")
		}

		var req SignUpRequest
		if err := c.Bind(&req); err != nil {
			log.Debug("failed to bind sign up request", "error", err)
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		if req.Name == "" || req.Email == "" || req.Password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "name, email, and password are required")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to hash password", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
		}

		user := types.User{
			Name:     req.Name,
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		if err := db.Create(&user).Error; err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return echo.NewHTTPError(http.StatusConflict, "email already exists")
			}
			log.Error("failed to create user", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
		}

		token, expiresAt, err := generateToken(user, cfg.CookieSecret)
		if err != nil {
			log.Error("failed to generate token", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
		}

		user.Password = ""
		log.Info("user signed up", "email", user.Email)
		return c.JSON(http.StatusCreated, AuthResponse{
			Token:     token,
			ExpiresAt: expiresAt,
			User:      user,
		})
	}
}

func RefreshToken(db *gorm.DB, cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)

		var user types.User
		if err := db.First(&user, userID).Error; err != nil {
			log.Debug("user not found for token refresh", "user_id", userID)
			return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
		}

		token, expiresAt, err := generateToken(user, cfg.CookieSecret)
		if err != nil {
			log.Error("failed to generate refresh token", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
		}

		user.Password = ""
		return c.JSON(http.StatusOK, AuthResponse{
			Token:     token,
			ExpiresAt: expiresAt,
			User:      user,
		})
	}
}

func GetUserIDFromContext(c echo.Context) uint {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return 0
	}
	return userID
}
