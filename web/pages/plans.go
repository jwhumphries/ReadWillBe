//go:build wasm

package pages

import (
	"net/url"
	"strconv"
	"strings"

	"readwillbe/types"
	"readwillbe/web/api"
	"readwillbe/web/components"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type PlansList struct {
	app.Compo
	user    *types.User
	plans   []types.Plan
	loading bool
	errMsg  string
}

func (p *PlansList) OnNav(ctx app.Context, u *url.URL) {
	if !api.IsAuthenticated() {
		ctx.Navigate("/auth/sign-in")
		return
	}
	p.loadData(ctx)
}

func (p *PlansList) loadData(ctx app.Context) {
	p.loading = true

	ctx.Async(func() {
		user, userErr := api.GetCurrentUser()
		plans, plansErr := api.GetPlans()

		ctx.Dispatch(func(ctx app.Context) {
			p.loading = false
			if userErr != nil || plansErr != nil {
				p.errMsg = "Failed to load plans"
				return
			}
			p.user = user
			p.plans = plans
		})
	})
}

func (p *PlansList) Render() app.UI {
	layout := &components.Layout{
		Title: "Plans - ReadWillBe",
		User:  p.user,
	}

	if p.loading {
		layout.Children = []app.UI{components.LoadingSpinner()}
		return layout
	}

	layout.Children = []app.UI{
		app.Div().Body(
			app.Div().Class("flex items-center justify-between mb-6").Body(
				app.H2().Class("text-3xl font-bold").Text("Reading Plans"),
				app.A().Href("/plans/create").Class("btn btn-primary").Text("Create Plan"),
			),
			app.If(len(p.plans) == 0, func() app.UI {
				return app.Div().Class("alert").Body(
					components.InfoIcon(),
					app.Div().Body(
						app.Div().Class("text-lg").Text("No reading plans yet"),
						app.Div().Class("text-sm").Text("Create your first reading plan to get started"),
					),
				)
			}).Else(func() app.UI {
				return app.Div().Class("grid gap-4 md:grid-cols-2 lg:grid-cols-3").Body(
					app.Range(p.plans).Slice(func(i int) app.UI {
						return &components.PlanCard{
							Plan:     p.plans[i],
							OnDelete: func() { p.loadData(app.Context{}) },
						}
					}),
				)
			}),
		),
	}

	return layout
}

type CreatePlan struct {
	app.Compo
	user    *types.User
	title   string
	errMsg  string
	loading bool
}

func (c *CreatePlan) OnNav(ctx app.Context, u *url.URL) {
	if !api.IsAuthenticated() {
		ctx.Navigate("/auth/sign-in")
		return
	}
	c.loadUser(ctx)
}

func (c *CreatePlan) loadUser(ctx app.Context) {
	ctx.Async(func() {
		user, _ := api.GetCurrentUser()
		ctx.Dispatch(func(ctx app.Context) {
			c.user = user
		})
	})
}

func (c *CreatePlan) Render() app.UI {
	layout := &components.Layout{
		Title: "Create Plan - ReadWillBe",
		User:  c.user,
	}

	layout.Children = []app.UI{
		app.Div().Body(
			app.H2().Class("text-3xl font-bold mb-6").Text("Create New Plan"),
			app.Div().Class("card bg-base-200 shadow-xl max-w-xl").Body(
				app.Div().Class("card-body").Body(
					app.If(c.errMsg != "", func() app.UI {
						return app.Div().Class("alert alert-error mb-4").Body(
							app.Span().Text(c.errMsg),
						)
					}),
					app.Form().
						ID("create-plan-form").
						Attr("enctype", "multipart/form-data").
						Attr("action", "/api/plans").
						Attr("method", "POST").
						Body(
							app.Div().Class("form-control").Body(
								app.Label().Class("label").Body(
									app.Span().Class("label-text").Text("Plan Title"),
								),
								app.Input().
									Type("text").
									Name("title").
									Class("input input-bordered").
									Placeholder("My Reading Plan").
									Value(c.title).
									OnInput(c.ValueTo(&c.title)).
									Required(true),
							),
							app.Div().Class("form-control mt-4").Body(
								app.Label().Class("label").Body(
									app.Span().Class("label-text").Text("CSV File"),
								),
								app.Input().
									Type("file").
									Name("csv").
									ID("csv-file").
									Class("file-input file-input-bordered w-full").
									Attr("accept", ".csv").
									Required(true),
							),
							app.Div().Class("alert alert-info mt-4").Body(
								components.InfoIcon(),
								app.Div().Body(
									app.P().Class("font-semibold").Text("CSV Format"),
									app.P().Class("text-sm").Text("Date,Reading Content"),
									app.P().Class("text-sm").Text("Supports: YYYY-MM-DD, MM/DD/YYYY, January 2024, 2024-W01"),
								),
							),
							app.Div().Class("card-actions justify-end mt-6").Body(
								app.A().Href("/plans").Class("btn btn-ghost").Text("Cancel"),
								app.Button().
									Type("submit").
									Class("btn btn-primary").
									Disabled(c.loading).
									Body(
										app.If(c.loading, func() app.UI {
											return app.Span().Class("loading loading-spinner")
										}),
										app.Text("Create Plan"),
									),
							),
						),
				),
			),
		),
	}

	return layout
}

type EditPlan struct {
	app.Compo
	user    *types.User
	plan    *types.Plan
	loading bool
	errMsg  string
	planID  uint
}

func (e *EditPlan) OnNav(ctx app.Context, u *url.URL) {
	if !api.IsAuthenticated() {
		ctx.Navigate("/auth/sign-in")
		return
	}

	parts := strings.Split(u.Path, "/")
	for i, part := range parts {
		if part == "plans" && i+1 < len(parts) {
			if id, err := strconv.ParseUint(parts[i+1], 10, 32); err == nil {
				e.planID = uint(id)
				break
			}
		}
	}

	e.loadData(ctx)
}

func (e *EditPlan) loadData(ctx app.Context) {
	e.loading = true

	ctx.Async(func() {
		user, _ := api.GetCurrentUser()
		plan, planErr := api.GetPlan(e.planID)

		ctx.Dispatch(func(ctx app.Context) {
			e.loading = false
			e.user = user
			if planErr != nil {
				e.errMsg = "Failed to load plan"
				return
			}
			e.plan = plan
		})
	})
}

func (e *EditPlan) Render() app.UI {
	layout := &components.Layout{
		Title: "Edit Plan - ReadWillBe",
		User:  e.user,
	}

	if e.loading {
		layout.Children = []app.UI{components.LoadingSpinner()}
		return layout
	}

	if e.errMsg != "" {
		layout.Children = []app.UI{
			app.Div().Class("alert alert-error").Body(
				app.Span().Text(e.errMsg),
			),
		}
		return layout
	}

	if e.plan == nil {
		layout.Children = []app.UI{
			app.Div().Class("alert alert-error").Body(
				app.Span().Text("Plan not found"),
			),
		}
		return layout
	}

	layout.Children = []app.UI{
		app.Div().Body(
			app.Div().Class("flex items-center justify-between mb-6").Body(
				app.H2().Class("text-3xl font-bold").Text("Edit Plan"),
				app.A().Href("/plans").Class("btn btn-ghost").Text("Back to Plans"),
			),
			app.Div().Class("card bg-base-200 shadow-xl").Body(
				app.Div().Class("card-body").Body(
					app.H3().Class("card-title").Text(e.plan.Title),
					app.Div().Class("overflow-x-auto").Body(
						app.Table().Class("table").Body(
							app.Thead().Body(
								app.Tr().Body(
									app.Th().Text("Date"),
									app.Th().Text("Content"),
									app.Th().Text("Status"),
								),
							),
							app.Tbody().Body(
								app.Range(e.plan.Readings).Slice(func(i int) app.UI {
									r := e.plan.Readings[i]
									statusBadge := "badge-ghost"
									if r.Status == types.StatusCompleted {
										statusBadge = "badge-success"
									} else if r.IsOverdue() {
										statusBadge = "badge-error"
									}
									return app.Tr().Body(
										app.Td().Text(r.Date.Format("Jan 2, 2006")),
										app.Td().Text(r.Content),
										app.Td().Body(
											app.Span().Class("badge "+statusBadge).Text(string(r.Status)),
										),
									)
								}),
							),
						),
					),
				),
			),
		),
	}

	return layout
}
