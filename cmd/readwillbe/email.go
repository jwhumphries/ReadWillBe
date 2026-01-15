package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	mail "github.com/wneessen/go-mail"
	"readwillbe/types"
)

type EmailService interface {
	SendDailyDigest(user types.User, readings []types.Reading, hostname string) error
	SendTestEmail(to, hostname string) error
}

func NewEmailService(cfg types.Config) EmailService {
	switch cfg.EmailProvider {
	case "smtp":
		return &SMTPEmailService{cfg: cfg}
	case "resend":
		return &ResendEmailService{cfg: cfg}
	default:
		return nil
	}
}

type SMTPEmailService struct {
	cfg types.Config
}

func (s *SMTPEmailService) SendDailyDigest(user types.User, readings []types.Reading, hostname string) error {
	html, text := renderDailyDigestEmail(user, readings, hostname)
	return s.send(user.GetNotificationEmail(), "Your readings for today", html, text)
}

func (s *SMTPEmailService) SendTestEmail(to, hostname string) error {
	html, text := renderTestEmail(hostname)
	return s.send(to, "Test Email from ReadWillBe", html, text)
}

func (s *SMTPEmailService) send(to, subject, htmlBody, textBody string) error {
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

type ResendEmailService struct {
	cfg types.Config
}

func (r *ResendEmailService) SendDailyDigest(user types.User, readings []types.Reading, hostname string) error {
	html, text := renderDailyDigestEmail(user, readings, hostname)
	return r.send(user.GetNotificationEmail(), "Your readings for today", html, text)
}

func (r *ResendEmailService) SendTestEmail(to, hostname string) error {
	html, text := renderTestEmail(hostname)
	return r.send(to, "Test Email from ReadWillBe", html, text)
}

func (r *ResendEmailService) send(to, subject, htmlBody, textBody string) error {
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
