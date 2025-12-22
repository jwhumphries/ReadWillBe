//go:build wasm

package components

import (
	"readwillbe/types"
	"readwillbe/web/api"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type Layout struct {
	app.Compo
	Title    string
	User     *types.User
	Children []app.UI
}

func (l *Layout) Render() app.UI {
	if l.User != nil && l.User.ID > 0 {
		return l.renderAuthenticated()
	}
	return l.renderUnauthenticated()
}

func (l *Layout) renderAuthenticated() app.UI {
	initial := ""
	if l.User != nil && len(l.User.Name) > 0 {
		initial = string([]rune(l.User.Name)[0:1])
	}

	return app.Div().Class("drawer lg:drawer-open").Body(
		app.Input().ID("drawer").Type("checkbox").Class("drawer-toggle"),
		app.Div().Class("drawer-content flex flex-col").Body(
			app.Header().Class("navbar bg-base-200 border-b border-base-300").Body(
				app.Div().Class("flex-none lg:hidden").Body(
					app.Label().For("drawer").Class("btn btn-square btn-ghost").Body(
						MenuIcon(),
					),
				),
				app.Div().Class("flex-1 text-center lg:text-left").Body(
					app.A().Href("/dashboard").Class("text-3xl font-bold link link-hover").Text("ReadWillBe"),
				),
			),
			app.Main().Class("flex-1 p-8 bg-base-100").Body(
				l.Children...,
			),
			app.Footer().Class("footer footer-center p-4 bg-base-200 text-base-content border-t border-base-300").Body(
				app.Aside().Body(
					app.P().Class("text-sm").Text("ReadWillBe PWA"),
				),
			),
		),
		app.Div().Class("drawer-side").Body(
			app.Label().For("drawer").Class("drawer-overlay"),
			app.Div().Class("w-64 min-h-full bg-base-200 flex flex-col").Body(
				app.Div().Class("p-4 text-center border-b border-base-300").Body(
					app.Div().Class("avatar placeholder").Body(
						app.Div().Class("bg-neutral text-neutral-content w-12 rounded-full").Body(
							app.Span().Class("text-xl").Text(initial),
						),
					),
					app.H2().Class("text-lg font-bold mt-2").Text(l.User.Name),
					app.P().Class("text-sm opacity-70").Text(l.User.Email),
				),
				app.Ul().Class("menu flex-1").Body(
					app.Li().Body(
						app.A().Href("/dashboard").Body(
							DashboardIcon(),
							app.Text("Dashboard"),
						),
					),
					app.Li().Body(
						app.A().Href("/history").Body(
							HistoryIcon(),
							app.Text("History"),
						),
					),
					app.Li().Body(
						app.A().Href("/plans").Body(
							PlansIcon(),
							app.Text("Plans"),
						),
					),
				),
				app.Ul().Class("menu").Body(
					app.Li().Body(
						app.A().Href("/account").Body(
							SettingsIcon(),
							app.Text("Settings"),
						),
					),
					app.Li().Body(
						app.A().Href("#").OnClick(l.handleSignOut).Body(
							SignOutIcon(),
							app.Text("Sign Out"),
						),
					),
				),
			),
		),
	)
}

func (l *Layout) renderUnauthenticated() app.UI {
	return app.Main().Class("flex-1 flex flex-col").Body(
		l.Children...,
	)
}

func (l *Layout) handleSignOut(ctx app.Context, e app.Event) {
	e.PreventDefault()
	api.ClearToken()
	ctx.Navigate("/auth/sign-in")
}

func LoadingSpinner() app.UI {
	return app.Div().Class("flex items-center justify-center py-12").Body(
		app.Span().Class("loading loading-spinner loading-lg"),
	)
}
