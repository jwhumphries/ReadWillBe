//go:build wasm

package main

import (
	"readwillbe/web/pages"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

func main() {
	app.Route("/", func() app.Composer { return &pages.Dashboard{} })
	app.Route("/dashboard", func() app.Composer { return &pages.Dashboard{} })
	app.Route("/auth/sign-in", func() app.Composer { return &pages.SignIn{} })
	app.Route("/auth/sign-up", func() app.Composer { return &pages.SignUp{} })
	app.Route("/history", func() app.Composer { return &pages.History{} })
	app.Route("/plans", func() app.Composer { return &pages.PlansList{} })
	app.Route("/plans/create", func() app.Composer { return &pages.CreatePlan{} })
	app.Route("/plans/{id}/edit", func() app.Composer { return &pages.EditPlan{} })
	app.Route("/account", func() app.Composer { return &pages.Account{} })

	app.RunWhenOnBrowser()
}
