//go:build wasm

package components

import (
	"fmt"

	"readwillbe/types"
	"readwillbe/web/api"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

type ReadingCard struct {
	app.Compo
	Reading    types.Reading
	IsOverdue  bool
	OnComplete func()
}

func (r *ReadingCard) Render() app.UI {
	cardClass := "card bg-base-200 shadow-xl"
	if r.IsOverdue {
		cardClass += " card-border border-error"
	}

	return app.Div().Class(cardClass).Body(
		app.Div().Class("card-body").Body(
			app.Div().Class("flex items-center gap-2 mb-2").Body(
				app.H3().Class("card-title flex-1").Text(r.Reading.Plan.Title),
				app.If(r.IsOverdue, func() app.UI {
					return app.Div().Class("badge badge-error gap-2").Body(
						OverdueIcon(),
						app.Text("Overdue"),
					)
				}),
			),
			app.P().Class("leading-relaxed").Text(r.Reading.Content),
			app.Div().Class("flex items-center justify-between mt-4").Body(
				app.Div().Class("flex items-center gap-2 text-sm opacity-75").Body(
					CalendarIcon(),
					app.Span().Text(r.Reading.Date.Format("January 2, 2006")),
				),
				app.Button().
					Class("btn btn-primary btn-sm").
					OnClick(r.handleComplete).
					Body(
						CheckIcon(),
						app.Text("Complete"),
					),
			),
		),
	)
}

func (r *ReadingCard) handleComplete(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		_, err := api.CompleteReading(r.Reading.ID)
		if err == nil && r.OnComplete != nil {
			ctx.Dispatch(func(ctx app.Context) {
				r.OnComplete()
			})
		}
	})
}

type CompletedReadingCard struct {
	app.Compo
	Reading      types.Reading
	OnUncomplete func()
}

func (c *CompletedReadingCard) Render() app.UI {
	return app.Div().Class("card bg-base-200 shadow-xl").Body(
		app.Div().Class("card-body").Body(
			app.H3().Class("card-title").Text(c.Reading.Plan.Title),
			app.P().Class("leading-relaxed").Text(c.Reading.Content),
			app.Div().Class("flex items-center gap-4 mt-4 text-sm").Body(
				app.Div().Class("badge badge-outline gap-1").Body(
					CalendarIcon(),
					app.Text(" Scheduled: "+c.Reading.Date.Format("Jan 2, 2006")),
				),
				app.If(c.Reading.CompletedAt != nil, func() app.UI {
					return app.Div().Class("badge badge-success gap-1").Body(
						CheckIcon(),
						app.Text(" Completed: "+c.Reading.CompletedAt.Format("Jan 2, 2006")),
					)
				}),
			),
		),
	)
}

type PlanCard struct {
	app.Compo
	Plan     types.Plan
	OnDelete func()
}

func (p *PlanCard) Render() app.UI {
	readingCount := len(p.Plan.Readings)
	completedCount := 0
	for _, r := range p.Plan.Readings {
		if r.Status == types.StatusCompleted {
			completedCount++
		}
	}

	return app.Div().Class("card bg-base-200 shadow-xl").Body(
		app.Div().Class("card-body").Body(
			app.H3().Class("card-title").Text(p.Plan.Title),
			app.Div().Class("stats stats-vertical lg:stats-horizontal shadow").Body(
				app.Div().Class("stat").Body(
					app.Div().Class("stat-title").Text("Total Readings"),
					app.Div().Class("stat-value").Text(fmt.Sprintf("%d", readingCount)),
				),
				app.Div().Class("stat").Body(
					app.Div().Class("stat-title").Text("Completed"),
					app.Div().Class("stat-value text-success").Text(fmt.Sprintf("%d", completedCount)),
				),
			),
			app.Div().Class("card-actions justify-end mt-4").Body(
				app.A().Href(fmt.Sprintf("/plans/%d/edit", p.Plan.ID)).Class("btn btn-sm btn-outline").Text("Edit"),
				app.Button().Class("btn btn-sm btn-error btn-outline").OnClick(p.handleDelete).Text("Delete"),
			),
		),
	)
}

func (p *PlanCard) handleDelete(ctx app.Context, e app.Event) {
	ctx.Async(func() {
		err := api.DeletePlan(p.Plan.ID)
		if err == nil && p.OnDelete != nil {
			ctx.Dispatch(func(ctx app.Context) {
				p.OnDelete()
			})
		}
	})
}
