//go:build wasm

package pages

import (
	"net/url"

	"readwillbe/types"
	"readwillbe/web/api"
	"readwillbe/web/components"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type History struct {
	app.Compo
	user     *types.User
	readings []types.Reading
	loading  bool
	errMsg   string
}

func (h *History) OnNav(ctx app.Context, u *url.URL) {
	if !api.IsAuthenticated() {
		ctx.Navigate("/auth/sign-in")
		return
	}
	h.loadData(ctx)
}

func (h *History) loadData(ctx app.Context) {
	h.loading = true

	ctx.Async(func() {
		user, userErr := api.GetCurrentUser()
		readings, readErr := api.GetHistory()

		ctx.Dispatch(func(ctx app.Context) {
			h.loading = false
			if userErr != nil || readErr != nil {
				h.errMsg = "Failed to load history"
				return
			}
			h.user = user
			h.readings = readings
		})
	})
}

func (h *History) Render() app.UI {
	layout := &components.Layout{
		Title: "History - ReadWillBe",
		User:  h.user,
	}

	if h.loading {
		layout.Children = []app.UI{components.LoadingSpinner()}
		return layout
	}

	layout.Children = []app.UI{
		app.Div().Body(
			app.H2().Class("text-3xl font-bold mb-6").Text("Reading History"),
			app.If(len(h.readings) == 0, func() app.UI {
				return app.Div().Class("alert").Body(
					components.InfoIcon(),
					app.Span().Text("No completed readings yet. Start reading to build your history!"),
				)
			}).Else(func() app.UI {
				return app.Div().Class("space-y-4").Body(
					app.Range(h.readings).Slice(func(i int) app.UI {
						return &components.CompletedReadingCard{Reading: h.readings[i]}
					}),
				)
			}),
		),
	}

	return layout
}
