package model

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// cookieSecretFromViper runs ConfigFromViper with only cookie_secret set,
// which is the sole required field, and returns the validation error.
func cookieSecretFromViper(t *testing.T, secret string) error {
	t.Helper()

	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("cookie_secret", secret)

	_, err := ConfigFromViper()
	return err
}

func TestConfigFromViper_CookieSecretAccepted(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{
			// `openssl rand -base64 32` — 44 chars, decodes to 32 bytes.
			name:   "base64 encoded 32 bytes",
			secret: "hHchpVEEQ8kFwvJyDPqvXQvJNBSPNCtaCFAJc8lHbTM=",
		},
		{
			// What the Helm chart's `randAlphaNum 32` produces. It happens to be
			// decodable as base64 (32 chars, all in the base64 alphabet) down to
			// only 24 bytes, but as a literal it carries ~192 bits of entropy.
			name:   "32 random alphanumeric chars",
			secret: "HLsQVaHTGWic37BxJlejV3bTk7iuO4cP",
		},
		{
			name:   "long passphrase with spaces and symbols",
			secret: "correct horse battery staple! plus some more words",
		},
		{
			name:   "exactly 32 chars of mixed case",
			secret: "aBcDeFgHiJkLmNoPqRsTuVwXyZaBcDeF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cookieSecretFromViper(t, tt.secret); err != nil {
				t.Errorf("expected secret to be accepted, got error: %v", err)
			}
		})
	}
}

func TestConfigFromViper_CookieSecretRejected(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		wantMsg string
	}{
		{
			name:    "empty",
			secret:  "",
			wantMsg: "cookie_secret is required",
		},
		{
			name:    "too short",
			secret:  "short",
			wantMsg: "at least 32 characters",
		},
		{
			// `openssl rand -base64 16` — 24 chars, under the length floor.
			name:    "base64 encoded 16 bytes",
			secret:  "F5kBqf7lS8XoRr2wYzN1Ag==",
			wantMsg: "at least 32 characters",
		},
		{
			name:    "31 chars is one short",
			secret:  strings.Repeat("aB3", 10) + "x",
			wantMsg: "at least 32 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cookieSecretFromViper(t, tt.secret)
			if err == nil {
				t.Fatalf("expected secret to be rejected, got nil error")
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestConfigFromViper_CookieSecretPreservedVerbatim(t *testing.T) {
	// The secret signs cookies as raw bytes; it must not be base64-decoded on
	// the way through, or every existing session breaks.
	const secret = "hHchpVEEQ8kFwvJyDPqvXQvJNBSPNCtaCFAJc8lHbTM="

	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("cookie_secret", secret)

	cfg, err := ConfigFromViper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(cfg.CookieSecret) != secret {
		t.Errorf("CookieSecret = %q, want %q", cfg.CookieSecret, secret)
	}
}

func TestConfig_BaseURL(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		want     string
	}{
		{
			// The form the Docker and Helm docs tell you to use.
			name:     "https URL is used as given",
			hostname: "https://read.example.com",
			want:     "https://read.example.com",
		},
		{
			// The Helm chart's default, and how the dev server is reached.
			name:     "http URL keeps its scheme",
			hostname: "http://localhost:8080",
			want:     "http://localhost:8080",
		},
		{
			// What the code required before base URLs were accepted, so it has
			// to keep resolving the same way.
			name:     "bare host gains an https scheme",
			hostname: "read.example.com",
			want:     "https://read.example.com",
		},
		{
			name:     "bare host with port gains an https scheme",
			hostname: "read.example.com:8443",
			want:     "https://read.example.com:8443",
		},
		{
			name:     "trailing slash is dropped",
			hostname: "https://read.example.com/",
			want:     "https://read.example.com",
		},
		{
			name:     "surrounding whitespace is dropped",
			hostname: "  https://read.example.com  ",
			want:     "https://read.example.com",
		},
		{
			name:     "scheme is recognised case-insensitively",
			hostname: "HTTPS://read.example.com",
			want:     "HTTPS://read.example.com",
		},
		{
			name:     "empty hostname stays empty",
			hostname: "",
			want:     "",
		},
		{
			name:     "whitespace-only hostname stays empty",
			hostname: "   ",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Config{Hostname: tt.hostname}.BaseURL()
			if got != tt.want {
				t.Errorf("BaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

// The scheme must not be doubled up, which is what made every digest link
// read https://http://localhost:8080/dashboard.
func TestConfig_BaseURLNeverDoublesTheScheme(t *testing.T) {
	for _, hostname := range []string{"http://localhost:8080", "https://read.example.com", "read.example.com"} {
		got := Config{Hostname: hostname}.BaseURL()
		if strings.Count(got, "://") != 1 {
			t.Errorf("BaseURL() = %q, want exactly one scheme separator", got)
		}
	}
}
