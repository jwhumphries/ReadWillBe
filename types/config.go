package types

import (
	"encoding/base64"
	"os"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const MinCookieSecretLength = 32
const MinCookieSecretEntropy = 128

type Config struct {
	DBPath          string
	CookieSecret    []byte
	AllowSignup     bool
	SeedDB          bool
	Port            string
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	Hostname        string

	// Email configuration (mutually exclusive: set EITHER SMTP OR Resend)
	EmailProvider string // "smtp" or "resend" (empty = disabled)

	// SMTP settings (used when EmailProvider = "smtp")
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string // "ReadWillBe <noreply@example.com>"
	SMTPTLS      string // "none", "starttls", "tls" (default: "starttls")

	// Resend settings (used when EmailProvider = "resend")
	ResendAPIKey string
	ResendFrom   string // "ReadWillBe <noreply@example.com>"
}

func (c Config) IsProduction() bool {
	env := strings.ToLower(os.Getenv("GO_ENV"))
	return env == "production" || env == "prod"
}

func (c Config) EmailEnabled() bool {
	return c.EmailProvider == "smtp" || c.EmailProvider == "resend"
}

func estimateEntropy(s string) int {
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, c := range s {
		switch {
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsDigit(c):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	charsetSize := 0
	if hasLower {
		charsetSize += 26
	}
	if hasUpper {
		charsetSize += 26
	}
	if hasDigit {
		charsetSize += 10
	}
	if hasSpecial {
		charsetSize += 32
	}

	if charsetSize == 0 {
		return 0
	}

	bitsPerChar := 0
	for charsetSize > 0 {
		bitsPerChar++
		charsetSize >>= 1
	}

	return len(s) * bitsPerChar
}

func ConfigFromViper() (Config, error) {
	cookieSecret := viper.GetString("cookie_secret")
	if cookieSecret == "" {
		return Config{}, errors.New("cookie_secret is required (set via READWILLBE_COOKIE_SECRET env var or config file)")
	}

	if len(cookieSecret) < MinCookieSecretLength {
		return Config{}, errors.Errorf("cookie_secret must be at least %d characters for security (generate with: openssl rand -base64 32)", MinCookieSecretLength)
	}

	if decoded, err := base64.StdEncoding.DecodeString(cookieSecret); err == nil {
		if len(decoded) < MinCookieSecretLength {
			return Config{}, errors.Errorf("cookie_secret base64-decoded length is only %d bytes, need at least %d bytes of entropy (generate with: openssl rand -base64 32)", len(decoded), MinCookieSecretLength)
		}
	} else {
		entropy := estimateEntropy(cookieSecret)
		if entropy < MinCookieSecretEntropy {
			return Config{}, errors.Errorf("cookie_secret has insufficient entropy (~%d bits, need %d+). Use a cryptographically random value (generate with: openssl rand -base64 32)", entropy, MinCookieSecretEntropy)
		}
	}

	port := viper.GetString("port")
	if port != "" && port[0] != ':' {
		port = ":" + port
	}

	// Validate email provider config
	emailProvider := strings.ToLower(viper.GetString("email_provider"))
	if emailProvider != "" && emailProvider != "smtp" && emailProvider != "resend" {
		return Config{}, errors.New("email_provider must be 'smtp', 'resend', or empty")
	}

	if emailProvider == "smtp" {
		if viper.GetString("smtp_host") == "" {
			return Config{}, errors.New("smtp_host is required when email_provider is 'smtp'")
		}
		if viper.GetString("smtp_from") == "" {
			return Config{}, errors.New("smtp_from is required when email_provider is 'smtp'")
		}
	}

	if emailProvider == "resend" {
		if viper.GetString("resend_api_key") == "" {
			return Config{}, errors.New("resend_api_key is required when email_provider is 'resend'")
		}
		if viper.GetString("resend_from") == "" {
			return Config{}, errors.New("resend_from is required when email_provider is 'resend'")
		}
	}

	smtpTLS := strings.ToLower(viper.GetString("smtp_tls"))
	if smtpTLS == "" {
		smtpTLS = "starttls"
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
		EmailProvider:   emailProvider,
		SMTPHost:        viper.GetString("smtp_host"),
		SMTPPort:        viper.GetInt("smtp_port"),
		SMTPUsername:    viper.GetString("smtp_username"),
		SMTPPassword:    viper.GetString("smtp_password"),
		SMTPFrom:        viper.GetString("smtp_from"),
		SMTPTLS:         smtpTLS,
		ResendAPIKey:    viper.GetString("resend_api_key"),
		ResendFrom:      viper.GetString("resend_from"),
	}, nil
}
