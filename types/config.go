package types

import (
	"os"
	"strconv"

	"github.com/pkg/errors"
)

type Config struct {
	DBPath       string
	CookieSecret []byte
	AllowSignup  bool
	Port         string
}

func ConfigFromEnv() (Config, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./tmp/readwillbe.db"
	}

	cookieSecret := os.Getenv("COOKIE_SECRET")
	if cookieSecret == "" {
		return Config{}, errors.New("COOKIE_SECRET env var is required")
	}

	allowSignup := true
	if val := os.Getenv("ALLOW_SIGNUP"); val != "" {
		parsed, err := strconv.ParseBool(val)
		if err != nil {
			return Config{}, errors.Wrap(err, "parsing ALLOW_SIGNUP")
		}
		allowSignup = parsed
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		DBPath:       dbPath,
		CookieSecret: []byte(cookieSecret),
		AllowSignup:  allowSignup,
		Port:         ":" + port,
	}, nil
}
