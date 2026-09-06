// Package email sends ReadWillBe notification email via the configured
// provider (SMTP or Resend).
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"readwillbe/internal/model"

	mail "github.com/wneessen/go-mail"
)

// sendTimeout bounds a single delivery attempt. The notification worker sends
// synchronously while iterating users, so an unreachable mail host must not
// stall the rest of the run.
const sendTimeout = 30 * time.Second

// resendAPIURL is the Resend transactional send endpoint.
const resendAPIURL = "https://api.resend.com/emails"

// Service is implemented by every supported email backend.
type Service interface {
	SendDailyDigest(user model.User, readings []model.Reading) error
	SendTestEmail(to string) error
}

// NewService returns the email service implementation matching
// cfg.EmailProvider, or an error if the provider is unset or unrecognised.
func NewService(cfg model.Config) (Service, error) {
	switch cfg.EmailProvider {
	case "smtp":
		return &SMTPService{cfg: cfg}, nil
	case "resend":
		return &ResendService{cfg: cfg, client: &http.Client{Timeout: sendTimeout}}, nil
	case "":
		return nil, fmt.Errorf("no email provider configured")
	default:
		return nil, fmt.Errorf("unknown email provider %q", cfg.EmailProvider)
	}
}

// SMTPService delivers email through an SMTP server.
type SMTPService struct {
	cfg model.Config
}

// SendDailyDigest renders and sends the daily reading digest for user.
func (s *SMTPService) SendDailyDigest(user model.User, readings []model.Reading) error {
	html, text := RenderDailyDigestEmail(user, readings, s.cfg.BaseURL())
	return s.send(user.GetNotificationEmail(), "Your readings for today", html, text)
}

// SendTestEmail sends a short test message to the given address.
func (s *SMTPService) SendTestEmail(to string) error {
	html, text := RenderTestEmail()
	return s.send(to, "Test Email from ReadWillBe", html, text)
}

func (s *SMTPService) send(to, subject, htmlBody, textBody string) error {
	m := mail.NewMsg()
	if err := m.From(s.cfg.SMTPFrom); err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}
	if err := m.To(to); err != nil {
		return fmt.Errorf("invalid to address: %w", err)
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextPlain, textBody)
	m.AddAlternativeString(mail.TypeTextHTML, htmlBody)

	var tlsPolicy mail.TLSPolicy
	switch s.cfg.SMTPTLS {
	case "none":
		tlsPolicy = mail.NoTLS
	case "tls":
		tlsPolicy = mail.TLSMandatory
	default:
		tlsPolicy = mail.TLSOpportunistic
	}

	opts := []mail.Option{
		mail.WithPort(s.cfg.SMTPPort),
		mail.WithTLSPolicy(tlsPolicy),
		mail.WithTimeout(sendTimeout),
	}

	if s.cfg.SMTPUsername != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(s.cfg.SMTPUsername),
			mail.WithPassword(s.cfg.SMTPPassword),
		)
	}

	c, err := mail.NewClient(s.cfg.SMTPHost, opts...)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}

	// WithTimeout bounds each network operation; the context bounds the whole
	// dial-and-send so a server that trickles responses still gives up.
	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()

	return c.DialAndSendWithContext(ctx, m)
}

// ResendService delivers email through the Resend HTTP API.
type ResendService struct {
	cfg    model.Config
	client *http.Client
}

// SendDailyDigest renders and sends the daily reading digest for user.
func (r *ResendService) SendDailyDigest(user model.User, readings []model.Reading) error {
	html, text := RenderDailyDigestEmail(user, readings, r.cfg.BaseURL())
	return r.send(user.GetNotificationEmail(), "Your readings for today", html, text)
}

// SendTestEmail sends a short test message to the given address.
func (r *ResendService) SendTestEmail(to string) error {
	html, text := RenderTestEmail()
	return r.send(to, "Test Email from ReadWillBe", html, text)
}

// resendPayload is the request body accepted by the Resend send endpoint.
type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
}

// buildResendPayload encodes a Resend send request. Bodies carry user-supplied
// plan titles and reading content verbatim, so they are JSON-encoded rather
// than interpolated into a format string.
func buildResendPayload(from, to, subject, htmlBody, textBody string) ([]byte, error) {
	return json.Marshal(resendPayload{
		From:    from,
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
		Text:    textBody,
	})
}

func (r *ResendService) send(to, subject, htmlBody, textBody string) error {
	payload, err := buildResendPayload(r.cfg.ResendFrom, to, subject, htmlBody, textBody)
	if err != nil {
		return fmt.Errorf("failed to encode resend payload: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := r.client
	if client == nil {
		client = &http.Client{Timeout: sendTimeout}
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}
	return nil
}
