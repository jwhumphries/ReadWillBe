package types

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	DBPath       string
	CookieSecret []byte
	AllowSignup  bool
	SeedDB       bool
	Port         string
}

func ConfigFromViper() (Config, error) {
	cookieSecret := viper.GetString("cookie_secret")
	if cookieSecret == "" {
		return Config{}, errors.New("cookie_secret is required (set via READWILLBE_COOKIE_SECRET env var or config file)")
	}

	port := viper.GetString("port")
	if port != "" && port[0] != ':' {
		port = ":" + port
	}

	return Config{
		DBPath:       viper.GetString("db_path"),
		CookieSecret: []byte(cookieSecret),
		AllowSignup:  viper.GetBool("allow_signup"),
		SeedDB:       viper.GetBool("seed_db"),
		Port:         port,
	}, nil
}

func ConfigFromEnv() (Config, error) {
	return Config{}, fmt.Errorf("ConfigFromEnv is deprecated, use ConfigFromViper instead")
}
