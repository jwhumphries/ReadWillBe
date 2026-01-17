package main

import (
	"bytes"
	"fmt"
	"html/template"

	"readwillbe/types"
)

const dailyDigestHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; background-color: #faf8f5; font-family: Georgia, 'Times New Roman', serif;">
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background-color: #faf8f5;">
    <tr>
      <td align="center" style="padding: 40px 20px;">
        <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width: 600px;">
          <!-- Header -->
          <tr>
            <td align="center" style="padding-bottom: 32px;">
              <h1 style="margin: 0; font-size: 32px; color: #3d3730;">ReadWillBe</h1>
            </td>
          </tr>

          <!-- Greeting -->
          <tr>
            <td style="padding-bottom: 24px;">
              <p style="margin: 0; font-size: 18px; color: #3d3730;">
                Hi {{.UserName}}, here are your readings for today:
              </p>
            </td>
          </tr>

          {{if .HasOverdue}}
          <!-- Overdue Warning -->
          <tr>
            <td style="padding-bottom: 16px;">
              <table role="presentation" width="100%" cellspacing="0" cellpadding="0"
                     style="background-color: #fef3cd; border-radius: 8px; border-left: 4px solid #e6c35c;">
                <tr>
                  <td style="padding: 12px 16px;">
                    <p style="margin: 0; color: #856404; font-size: 14px;">
                      You have {{.OverdueCount}} overdue reading(s)
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          {{end}}

          <!-- Reading Cards -->
          {{range .Readings}}
          <tr>
            <td style="padding-bottom: 16px;">
              <table role="presentation" width="100%" cellspacing="0" cellpadding="0"
                     style="background-color: #f0ede8; border-radius: 12px; {{if .IsOverdue}}border-left: 4px solid #c44536;{{end}}">
                <tr>
                  <td style="padding: 20px;">
                    <!-- Plan Title -->
                    <table role="presentation" width="100%" cellspacing="0" cellpadding="0">
                      <tr>
                        <td>
                          <h3 style="margin: 0 0 8px 0; font-size: 18px; color: #3d3730;">
                            {{.PlanTitle}}
                          </h3>
                        </td>
                        {{if .IsOverdue}}
                        <td align="right" style="vertical-align: top;">
                          <span style="display: inline-block; background-color: #c44536; color: white;
                                       font-size: 12px; padding: 4px 8px; border-radius: 4px;">
                            Overdue
                          </span>
                        </td>
                        {{end}}
                      </tr>
                    </table>
                    <!-- Reading Content -->
                    <p style="margin: 0 0 12px 0; font-size: 16px; color: #3d3730; line-height: 1.5;">
                      {{.Content}}
                    </p>
                    <!-- Date -->
                    <p style="margin: 0; font-size: 14px; color: #6b6560;">
                      {{.FormattedDate}}
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          {{end}}

          <!-- CTA Button -->
          <tr>
            <td align="center" style="padding-top: 16px;">
              <a href="{{.DashboardURL}}"
                 style="display: inline-block; background-color: #4a8c4a; color: white;
                        text-decoration: none; padding: 12px 32px; border-radius: 8px;
                        font-size: 16px; font-weight: bold;">
                View Dashboard
              </a>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td align="center" style="padding-top: 40px;">
              <p style="margin: 0; font-size: 12px; color: #6b6560;">
                You're receiving this because you enabled email notifications.
                <br>
                <a href="{{.SettingsURL}}" style="color: #4a8c4a;">Manage notification settings</a>
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

const dailyDigestTextTemplate = `ReadWillBe - Your Daily Readings

Hi {{.UserName}},

Here are your readings for today:
{{if .HasOverdue}}
You have {{.OverdueCount}} overdue reading(s)
{{end}}
{{range .Readings}}
---
{{.PlanTitle}}{{if .IsOverdue}} [OVERDUE]{{end}}
{{.Content}}
{{.FormattedDate}}
{{end}}
---

View your dashboard: {{.DashboardURL}}

Manage notifications: {{.SettingsURL}}`

const testEmailHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
</head>
<body style="margin: 0; padding: 40px; background-color: #faf8f5; font-family: Georgia, serif;">
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width: 600px; margin: 0 auto;">
    <tr>
      <td align="center" style="padding-bottom: 24px;">
        <h1 style="margin: 0; font-size: 32px; color: #3d3730;">ReadWillBe</h1>
      </td>
    </tr>
    <tr>
      <td style="background-color: #f0ede8; border-radius: 12px; padding: 32px; text-align: center;">
        <h2 style="margin: 0 0 16px 0; color: #4a8c4a;">Email Configuration Working!</h2>
        <p style="margin: 0; color: #3d3730; font-size: 16px;">
          Your email notifications are configured correctly.
        </p>
      </td>
    </tr>
  </table>
</body>
</html>`

const testEmailTextTemplate = `ReadWillBe - Test Email

Email Configuration Working!

Your email notifications are configured correctly.`

type emailReading struct {
	PlanTitle     string
	Content       string
	FormattedDate string
	IsOverdue     bool
}

type dailyDigestData struct {
	UserName     string
	Readings     []emailReading
	HasOverdue   bool
	OverdueCount int
	DashboardURL string
	SettingsURL  string
}

func renderDailyDigestEmail(user types.User, readings []types.Reading, hostname string) (html, text string) {
	data := dailyDigestData{
		UserName:     user.Name,
		DashboardURL: fmt.Sprintf("https://%s/dashboard", hostname),
		SettingsURL:  fmt.Sprintf("https://%s/account", hostname),
	}

	for _, r := range readings {
		isOverdue := r.IsOverdue()
		if isOverdue {
			data.HasOverdue = true
			data.OverdueCount++
		}
		data.Readings = append(data.Readings, emailReading{
			PlanTitle:     r.Plan.Title,
			Content:       r.Content,
			FormattedDate: r.Date.Format("January 2, 2006"),
			IsOverdue:     isOverdue,
		})
	}

	htmlTmpl := template.Must(template.New("html").Parse(dailyDigestHTMLTemplate))
	textTmpl := template.Must(template.New("text").Parse(dailyDigestTextTemplate))

	var htmlBuf, textBuf bytes.Buffer
	_ = htmlTmpl.Execute(&htmlBuf, data)
	_ = textTmpl.Execute(&textBuf, data)

	return htmlBuf.String(), textBuf.String()
}

func renderTestEmail(_ string) (html, text string) {
	return testEmailHTMLTemplate, testEmailTextTemplate
}
