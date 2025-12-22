//go:build wasm

package pages

import (
	"net/url"

	"readwillbe/web/api"
	"readwillbe/web/components"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type SignIn struct {
	app.Compo
	email    string
	password string
	errMsg   string
	loading  bool
}

func (s *SignIn) OnNav(ctx app.Context, u *url.URL) {
	if api.IsAuthenticated() {
		ctx.Navigate("/dashboard")
	}
}

func (s *SignIn) Render() app.UI {
	layout := &components.Layout{Title: "Sign In - ReadWillBe"}

	layout.Children = []app.UI{
		app.Div().Class("flex items-center justify-center min-h-screen bg-base-100").Body(
			app.Div().Class("card w-full max-w-md bg-base-200 shadow-xl").Body(
				app.Div().Class("card-body").Body(
					app.H1().Class("card-title text-3xl font-bold text-center justify-center mb-6").Text("ReadWillBe"),
					app.If(s.errMsg != "", func() app.UI {
						return app.Div().Class("alert alert-error mb-4").Body(
							app.Span().Text(s.errMsg),
						)
					}),
					app.Form().OnSubmit(s.handleSubmit).Body(
						app.Div().Class("form-control").Body(
							app.Label().Class("label").Body(
								app.Span().Class("label-text").Text("Email"),
							),
							app.Input().
								Type("email").
								Class("input input-bordered").
								Placeholder("email@example.com").
								Value(s.email).
								OnInput(s.ValueTo(&s.email)).
								Required(true),
						),
						app.Div().Class("form-control mt-4").Body(
							app.Label().Class("label").Body(
								app.Span().Class("label-text").Text("Password"),
							),
							app.Input().
								Type("password").
								Class("input input-bordered").
								Value(s.password).
								OnInput(s.ValueTo(&s.password)).
								Required(true),
						),
						app.Div().Class("form-control mt-6").Body(
							app.Button().
								Type("submit").
								Class("btn btn-primary").
								Disabled(s.loading).
								Body(
									app.If(s.loading, func() app.UI {
										return app.Span().Class("loading loading-spinner")
									}),
									app.Text("Sign In"),
								),
						),
					),
					app.Div().Class("divider"),
					app.Div().Class("text-center").Body(
						app.Text("Don't have an account? "),
						app.A().Href("/auth/sign-up").Class("link link-primary").Text("Sign Up"),
					),
				),
			),
		),
	}

	return layout
}

func (s *SignIn) handleSubmit(ctx app.Context, e app.Event) {
	e.PreventDefault()
	s.loading = true
	s.errMsg = ""

	ctx.Async(func() {
		resp, err := api.SignIn(s.email, s.password)
		ctx.Dispatch(func(ctx app.Context) {
			s.loading = false
			if err != nil {
				s.errMsg = "Invalid email or password"
				return
			}
			api.SetToken(resp.Token)
			ctx.Navigate("/dashboard")
		})
	})
}

type SignUp struct {
	app.Compo
	name     string
	email    string
	password string
	errMsg   string
	loading  bool
}

func (s *SignUp) OnNav(ctx app.Context, u *url.URL) {
	if api.IsAuthenticated() {
		ctx.Navigate("/dashboard")
	}
}

func (s *SignUp) Render() app.UI {
	layout := &components.Layout{Title: "Sign Up - ReadWillBe"}

	layout.Children = []app.UI{
		app.Div().Class("flex items-center justify-center min-h-screen bg-base-100").Body(
			app.Div().Class("card w-full max-w-md bg-base-200 shadow-xl").Body(
				app.Div().Class("card-body").Body(
					app.H1().Class("card-title text-3xl font-bold text-center justify-center mb-6").Text("Create Account"),
					app.If(s.errMsg != "", func() app.UI {
						return app.Div().Class("alert alert-error mb-4").Body(
							app.Span().Text(s.errMsg),
						)
					}),
					app.Form().OnSubmit(s.handleSubmit).Body(
						app.Div().Class("form-control").Body(
							app.Label().Class("label").Body(
								app.Span().Class("label-text").Text("Name"),
							),
							app.Input().
								Type("text").
								Class("input input-bordered").
								Placeholder("Your name").
								Value(s.name).
								OnInput(s.ValueTo(&s.name)).
								Required(true),
						),
						app.Div().Class("form-control mt-4").Body(
							app.Label().Class("label").Body(
								app.Span().Class("label-text").Text("Email"),
							),
							app.Input().
								Type("email").
								Class("input input-bordered").
								Placeholder("email@example.com").
								Value(s.email).
								OnInput(s.ValueTo(&s.email)).
								Required(true),
						),
						app.Div().Class("form-control mt-4").Body(
							app.Label().Class("label").Body(
								app.Span().Class("label-text").Text("Password"),
							),
							app.Input().
								Type("password").
								Class("input input-bordered").
								Value(s.password).
								OnInput(s.ValueTo(&s.password)).
								Required(true),
						),
						app.Div().Class("form-control mt-6").Body(
							app.Button().
								Type("submit").
								Class("btn btn-primary").
								Disabled(s.loading).
								Body(
									app.If(s.loading, func() app.UI {
										return app.Span().Class("loading loading-spinner")
									}),
									app.Text("Sign Up"),
								),
						),
					),
					app.Div().Class("divider"),
					app.Div().Class("text-center").Body(
						app.Text("Already have an account? "),
						app.A().Href("/auth/sign-in").Class("link link-primary").Text("Sign In"),
					),
				),
			),
		),
	}

	return layout
}

func (s *SignUp) handleSubmit(ctx app.Context, e app.Event) {
	e.PreventDefault()
	s.loading = true
	s.errMsg = ""

	ctx.Async(func() {
		resp, err := api.SignUp(s.name, s.email, s.password)
		ctx.Dispatch(func(ctx app.Context) {
			s.loading = false
			if err != nil {
				s.errMsg = "Failed to create account. Email may already be in use."
				return
			}
			api.SetToken(resp.Token)
			ctx.Navigate("/dashboard")
		})
	})
}
