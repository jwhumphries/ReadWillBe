package email

import (
	"strings"
	"testing"
	"time"

	"readwillbe/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dailyReading(planTitle, content string, date time.Time, status model.ReadingStatus) model.Reading {
	return model.Reading{
		Plan:     model.Plan{Title: planTitle},
		Date:     date,
		DateType: model.DateTypeDay,
		Content:  content,
		Status:   status,
	}
}

func TestRenderDailyDigestEmail_IncludesReadingsAndGreeting(t *testing.T) {
	user := model.User{Name: "Ada", Email: "ada@example.com"}
	readings := []model.Reading{
		dailyReading("War and Peace", "Chapter 1", time.Now(), model.StatusPending),
	}

	html, text := RenderDailyDigestEmail(user, readings, "https://read.example.com")

	for _, body := range []string{html, text} {
		assert.Contains(t, body, "Ada")
		assert.Contains(t, body, "War and Peace")
		assert.Contains(t, body, "Chapter 1")
	}
}

func TestRenderDailyDigestEmail_BuildsLinksFromBaseURL(t *testing.T) {
	html, text := RenderDailyDigestEmail(model.User{Name: "Ada"}, nil, "https://read.example.com")

	assert.Contains(t, html, "https://read.example.com/dashboard")
	assert.Contains(t, html, "https://read.example.com/account")
	assert.Contains(t, text, "https://read.example.com/dashboard")
}

// The base URL carries its own scheme, so the renderer must not add one.
func TestRenderDailyDigestEmail_PreservesBaseURLScheme(t *testing.T) {
	html, _ := RenderDailyDigestEmail(model.User{Name: "Ada"}, nil, "http://localhost:8080")

	assert.Contains(t, html, "http://localhost:8080/dashboard")
	assert.NotContains(t, html, "https://http://")
}

func TestRenderDailyDigestEmail_FlagsOverdueReadings(t *testing.T) {
	twoDaysAgo := time.Now().AddDate(0, 0, -2)
	readings := []model.Reading{
		dailyReading("War and Peace", "Chapter 1", twoDaysAgo, model.StatusPending),
		dailyReading("Moby Dick", "Chapter 2", time.Now(), model.StatusPending),
	}

	html, _ := RenderDailyDigestEmail(model.User{Name: "Ada"}, readings, "https://read.example.com")

	assert.Contains(t, html, "1 overdue reading(s)")
}

func TestRenderDailyDigestEmail_OmitsOverdueBannerWhenNoneOverdue(t *testing.T) {
	readings := []model.Reading{
		dailyReading("Moby Dick", "Chapter 2", time.Now(), model.StatusPending),
	}

	html, _ := RenderDailyDigestEmail(model.User{Name: "Ada"}, readings, "https://read.example.com")

	assert.NotContains(t, html, "overdue reading(s)")
}

// Plan titles and reading content are user-supplied, so they must not be able
// to inject markup into the HTML body.
func TestRenderDailyDigestEmail_EscapesUserSuppliedContent(t *testing.T) {
	readings := []model.Reading{
		dailyReading("<script>alert(1)</script>", "Chapter 1", time.Now(), model.StatusPending),
	}

	html, _ := RenderDailyDigestEmail(model.User{Name: "Ada"}, readings, "https://read.example.com")

	assert.NotContains(t, html, "<script>alert(1)</script>")
}

func TestRenderDailyDigestEmail_ProducesBothBodies(t *testing.T) {
	html, text := RenderDailyDigestEmail(model.User{Name: "Ada"}, nil, "https://read.example.com")

	require.NotEmpty(t, html)
	require.NotEmpty(t, text)
	assert.True(t, strings.HasPrefix(strings.TrimSpace(html), "<!DOCTYPE html>"))
	assert.NotContains(t, text, "<td")
}

func TestRenderTestEmail_ProducesBothBodies(t *testing.T) {
	html, text := RenderTestEmail()

	require.NotEmpty(t, html)
	require.NotEmpty(t, text)
	assert.Contains(t, html, "ReadWillBe")
	assert.Contains(t, text, "ReadWillBe")
}
