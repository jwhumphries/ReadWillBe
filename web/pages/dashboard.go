//go:build wasm

package pages

import (
	"fmt"
	"net/url"

	"readwillbe/types"
	"readwillbe/web/api"
	"readwillbe/web/components"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Dashboard struct {
	app.Compo
	user            *types.User
	todayReadings   []types.Reading
	overdueReadings []types.Reading
	loading         bool
	errMsg          string
}

func (d *Dashboard) OnNav(ctx app.Context, u *url.URL) {
	if !api.IsAuthenticated() {
		ctx.Navigate("/auth/sign-in")
		return
	}
	d.loadData(ctx)
}

func (d *Dashboard) loadData(ctx app.Context) {
	d.loading = true

	ctx.Async(func() {
		user, userErr := api.GetCurrentUser()
		dashboard, dashErr := api.GetDashboard()

		ctx.Dispatch(func(ctx app.Context) {
			d.loading = false
			if userErr != nil || dashErr != nil {
				d.errMsg = "Failed to load dashboard data"
				return
			}
			d.user = user
			d.todayReadings = dashboard.TodayReadings
			d.overdueReadings = dashboard.OverdueReadings
		})
	})
}

func (d *Dashboard) Render() app.UI {
	layout := &components.Layout{
		Title: "Dashboard - ReadWillBe",
		User:  d.user,
	}

	if d.loading {
		layout.Children = []app.UI{components.LoadingSpinner()}
		return layout
	}

	if d.errMsg != "" {
		layout.Children = []app.UI{
			app.Div().Class("alert alert-error").Body(
				app.Span().Text(d.errMsg),
			),
		}
		return layout
	}

	layout.Children = []app.UI{
		app.Div().Class("space-y-8").Body(
			app.Div().Body(
				app.H2().Class("text-3xl font-bold mb-6").Text("Today's Reading"),
				app.If(len(d.todayReadings) == 0, func() app.UI {
					return app.Div().Class("alert").Attr("role", "alert").Body(
						components.InfoIcon(),
						app.Div().Body(
							app.Div().Class("text-lg").Text("No readings scheduled for today"),
							app.Div().Class("text-sm").Body(
								app.Text("Enjoy your free time or check your "),
								app.A().Href("/plans").Class("link link-primary").Text("reading plans"),
							),
						),
					)
				}).Else(func() app.UI {
					return app.Div().Class("space-y-4").Body(
						app.Range(d.todayReadings).Slice(func(i int) app.UI {
							return &components.ReadingCard{
								Reading:    d.todayReadings[i],
								IsOverdue:  false,
								OnComplete: func() { d.loadData(app.Context{}) },
							}
						}),
					)
				}),
			),
			app.If(len(d.overdueReadings) > 0, func() app.UI {
				return app.Div().Body(
					app.Div().Class("alert alert-warning mb-4").Attr("role", "alert").Body(
						components.WarningIcon(),
						app.Span().Text(fmt.Sprintf("You have %d overdue reading(s)", len(d.overdueReadings))),
					),
					app.Div().Class("space-y-4").Body(
						app.Range(d.overdueReadings).Slice(func(i int) app.UI {
							return &components.ReadingCard{
								Reading:    d.overdueReadings[i],
								IsOverdue:  true,
								OnComplete: func() { d.loadData(app.Context{}) },
							}
						}),
					),
				)
			}),
		),
	}

	return layout
}
