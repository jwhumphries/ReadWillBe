//go:build wasm

package pages

import (
	"net/url"

	"readwillbe/types"
	"readwillbe/web/api"
	"readwillbe/web/components"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Account struct {
	app.Compo
	user                 *types.User
	notificationsEnabled bool
	notificationTime     string
	loading              bool
	saving               bool
	errMsg               string
	success              string
}

func (a *Account) OnNav(ctx app.Context, u *url.URL) {
	if !api.IsAuthenticated() {
		ctx.Navigate("/auth/sign-in")
		return
	}
	a.loadData(ctx)
}

func (a *Account) loadData(ctx app.Context) {
	a.loading = true

	ctx.Async(func() {
		user, err := api.GetCurrentUser()
		ctx.Dispatch(func(ctx app.Context) {
			a.loading = false
			if err != nil {
				a.errMsg = "Failed to load account data"
				return
			}
			a.user = user
			a.notificationsEnabled = user.NotificationsEnabled
			a.notificationTime = user.NotificationTime
		})
	})
}

func (a *Account) Render() app.UI {
	layout := &components.Layout{
		Title: "Account - ReadWillBe",
		User:  a.user,
	}

	if a.loading {
		layout.Children = []app.UI{components.LoadingSpinner()}
		return layout
	}

	layout.Children = []app.UI{
		app.Div().Body(
			app.H2().Class("text-3xl font-bold mb-6").Text("Account Settings"),
			app.Div().Class("card bg-base-200 shadow-xl mb-6").Body(
				app.Div().Class("card-body").Body(
					app.H3().Class("card-title").Text("Profile"),
					app.Ul().Class("list").Body(
						app.Li().Class("list-row").Body(
							app.Span().Class("font-semibold").Text("Name"),
							app.Span().Text(a.user.Name),
						),
						app.Li().Class("list-row").Body(
							app.Span().Class("font-semibold").Text("Email"),
							app.Span().Text(a.user.Email),
						),
					),
				),
			),
			app.Div().Class("card bg-base-200 shadow-xl").Body(
				app.Div().Class("card-body").Body(
					app.H3().Class("card-title").Text("Notifications"),
					app.If(a.success != "", func() app.UI {
						return app.Div().Class("alert alert-success mb-4").Body(
							app.Span().Text(a.success),
						)
					}),
					app.If(a.errMsg != "", func() app.UI {
						return app.Div().Class("alert alert-error mb-4").Body(
							app.Span().Text(a.errMsg),
						)
					}),
					app.Form().OnSubmit(a.handleSave).Body(
						app.Div().Class("form-control").Body(
							app.Label().Class("label cursor-pointer").Body(
								app.Span().Class("label-text").Text("Enable Notifications"),
								app.Input().
									Type("checkbox").
									Class("toggle toggle-primary").
									Checked(a.notificationsEnabled).
									OnChange(a.handleNotificationToggle),
							),
						),
						app.If(a.notificationsEnabled, func() app.UI {
							return app.Div().Class("form-control mt-4").Body(
								app.Label().Class("label").Body(
									app.Span().Class("label-text").Text("Notification Time"),
								),
								app.Input().
									Type("time").
									Class("input input-bordered").
									Value(a.notificationTime).
									OnInput(a.ValueTo(&a.notificationTime)),
							)
						}),
						app.Div().Class("card-actions justify-end mt-6").Body(
							app.Button().
								Type("submit").
								Class("btn btn-primary").
								Disabled(a.saving).
								Body(
									app.If(a.saving, func() app.UI {
										return app.Span().Class("loading loading-spinner")
									}),
									app.Text("Save Settings"),
								),
						),
					),
				),
			),
		),
	}

	return layout
}

func (a *Account) handleNotificationToggle(ctx app.Context, e app.Event) {
	a.notificationsEnabled = !a.notificationsEnabled
}

func (a *Account) handleSave(ctx app.Context, e app.Event) {
	e.PreventDefault()
	a.saving = true
	a.errMsg = ""
	a.success = ""

	ctx.Async(func() {
		user, err := api.UpdateSettings(a.notificationsEnabled, a.notificationTime)
		ctx.Dispatch(func(ctx app.Context) {
			a.saving = false
			if err != nil {
				a.errMsg = "Failed to save settings"
				return
			}
			a.user = user
			a.success = "Settings saved successfully"
		})
	})
}
