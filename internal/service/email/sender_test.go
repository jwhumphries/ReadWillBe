package email

import (
	"encoding/json"
	"testing"

	"readwillbe/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService_SMTPProvider(t *testing.T) {
	svc, err := NewService(model.Config{EmailProvider: "smtp"})

	require.NoError(t, err)
	assert.IsType(t, &SMTPService{}, svc)
}

func TestNewService_ResendProvider(t *testing.T) {
	svc, err := NewService(model.Config{EmailProvider: "resend"})

	require.NoError(t, err)
	assert.IsType(t, &ResendService{}, svc)
}

func TestNewService_UnknownProviderReturnsError(t *testing.T) {
	svc, err := NewService(model.Config{EmailProvider: "carrier-pigeon"})

	require.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewService_EmptyProviderReturnsError(t *testing.T) {
	svc, err := NewService(model.Config{})

	require.Error(t, err)
	assert.Nil(t, svc)
}

func TestBuildResendPayload_IncludesEnvelopeFields(t *testing.T) {
	payload, err := buildResendPayload("ReadWillBe <no-reply@example.com>", "reader@example.com",
		"Your readings for today", "<p>hi</p>", "hi")
	require.NoError(t, err)

	var decoded struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		HTML    string   `json:"html"`
		Text    string   `json:"text"`
	}
	require.NoError(t, json.Unmarshal(payload, &decoded))

	assert.Equal(t, "ReadWillBe <no-reply@example.com>", decoded.From)
	assert.Equal(t, []string{"reader@example.com"}, decoded.To)
	assert.Equal(t, "Your readings for today", decoded.Subject)
	assert.Equal(t, "<p>hi</p>", decoded.HTML)
	assert.Equal(t, "hi", decoded.Text)
}

// Reading content reaches the payload verbatim, so a control character in a
// plan title has to survive encoding. Go's %q verb renders one as \x01, which
// is not a legal JSON escape and makes Resend reject the whole request.
func TestBuildResendPayload_EncodesControlCharactersAsValidJSON(t *testing.T) {
	payload, err := buildResendPayload("ReadWillBe <no-reply@example.com>", "reader@example.com",
		"Your readings for today", "<p>Chapter\x01One</p>", "Chapter\x01One")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(payload, &decoded))

	assert.Equal(t, "<p>Chapter\x01One</p>", decoded["html"])
	assert.Equal(t, "Chapter\x01One", decoded["text"])
}

func TestBuildResendPayload_EncodesNonASCIIContent(t *testing.T) {
	payload, err := buildResendPayload("ReadWillBe <no-reply@example.com>", "reader@example.com",
		"Lectures d'aujourd'hui", "<p>Café — Chapitre «un»</p>", "Café — Chapitre «un»")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(payload, &decoded))

	assert.Equal(t, "Café — Chapitre «un»", decoded["text"])
	assert.Equal(t, "Lectures d'aujourd'hui", decoded["subject"])
}
