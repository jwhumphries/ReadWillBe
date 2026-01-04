package types

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const MinCookieSecretLength = 32

type Config struct {
	DBPath          string
	CookieSecret    []byte
	AllowSignup     bool
	SeedDB          bool
	Port            string
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	Hostname        string
}

func (c Config) IsProduction() bool {
	env := strings.ToLower(os.Getenv("GO_ENV"))
	return env == "production" || env == "prod"
}

func ConfigFromViper() (Config, error) {
	cookieSecret := viper.GetString("cookie_secret")
	if cookieSecret == "" {
		return Config{}, errors.New("cookie_secret is required (set via READWILLBE_COOKIE_SECRET env var or config file)")
	}

	if len(cookieSecret) < MinCookieSecretLength {
		return Config{}, errors.Errorf("cookie_secret must be at least %d characters for security", MinCookieSecretLength)
	}

	port := viper.GetString("port")
	if port != "" && port[0] != ':' {
		port = ":" + port
	}

	return Config{
		DBPath:          viper.GetString("db_path"),
		CookieSecret:    []byte(cookieSecret),
		AllowSignup:     viper.GetBool("allow_signup"),
		SeedDB:          viper.GetBool("seed_db"),
		Port:            port,
		VAPIDPublicKey:  viper.GetString("vapid_public_key"),
		VAPIDPrivateKey: viper.GetString("vapid_private_key"),
		Hostname:        viper.GetString("hostname"),
	}, nil
}
