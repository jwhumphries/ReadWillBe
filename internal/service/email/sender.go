package email

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	mail "github.com/wneessen/go-mail"
	"readwillbe/internal/model"
)

type Service interface {
	SendDailyDigest(user model.User, readings []model.Reading, hostname string) error
	SendTestEmail(to, hostname string) error
}

func NewService(cfg model.Config) Service {
	switch cfg.EmailProvider {
	case "smtp":
		return &SMTPService{cfg: cfg}
	case "resend":
		return &ResendService{cfg: cfg}
	default:
		return nil
	}
}

type SMTPService struct {
	cfg model.Config
}

func (s *SMTPService) SendDailyDigest(user model.User, readings []model.Reading, hostname string) error {
	html, text := RenderDailyDigestEmail(user, readings, hostname)
	return s.send(user.GetNotificationEmail(), "Your readings for today", html, text)
}

func (s *SMTPService) SendTestEmail(to, hostname string) error {
	html, text := RenderTestEmail(hostname)
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

	return c.DialAndSend(m)
}

type ResendService struct {
	cfg model.Config
}

func (r *ResendService) SendDailyDigest(user model.User, readings []model.Reading, hostname string) error {
	html, text := RenderDailyDigestEmail(user, readings, hostname)
	return r.send(user.GetNotificationEmail(), "Your readings for today", html, text)
}

func (r *ResendService) SendTestEmail(to, hostname string) error {
	html, text := RenderTestEmail(hostname)
	return r.send(to, "Test Email from ReadWillBe", html, text)
}

func (r *ResendService) send(to, subject, htmlBody, textBody string) error {
	payload := fmt.Sprintf(`{
		"from": %q,
		"to": [%q],
		"subject": %q,
		"html": %q,
		"text": %q
	}`, r.cfg.ResendFrom, to, subject, htmlBody, textBody)

	req, err := http.NewRequestWithContext(context.Background(), "POST", "https://api.resend.com/emails",
		strings.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}
	return nil
}
