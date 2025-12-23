package main

import (
	"context"
	"fmt"

	"dagger/readwillbe/internal/dagger"
)

type Readwillbe struct{}

// Build runs the full build pipeline: templ, css, lint, test, and build.
func (m *Readwillbe) Build(
	ctx context.Context,
	// The source directory
	source *dagger.Directory,
	// The version tag for the build (default: "dev")
	// +optional
	// +default="dev"
	version string,
) (*dagger.Container, error) {
	// 1. Frontend: Build CSS
	// Use the custom frontend image as requested
	frontend := dagger.Container().From("ghcr.io/jwhumphries/frontend:latest")

	// Mount source and run build
	// We only need input.css and package files effectively, but mounting source is easiest.
	cssDir := frontend.
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithExec([]string{"bun", "install"}).
		WithExec([]string{"bun", "run", "build"}).
		Directory("/app/static/css")

	// 2. Templ: Generate Go code
	// Use Go image
	gobase := dagger.Container().From("golang:1.25-alpine")

	// Install templ
	// We mount caches to speed up build
	templer := gobase.
		WithEnvVariable("GOCACHE", "/go-build-cache").
		WithEnvVariable("GOMODCACHE", "/go-mod-cache").
		WithMountedCache("/go-build-cache", dagger.CacheVolume("go-build-cache")).
		WithMountedCache("/go-mod-cache", dagger.CacheVolume("go-mod-cache")).
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		WithExec([]string{"go", "install", "github.com/a-h/templ/cmd/templ@latest"})

	// Generate templ files
	// We need to return the directory with generated files to be used by subsequent steps
	templSource := templer.
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithExec([]string{"templ", "generate"}).
		Directory("/app")

	// 3. Lint
	// Use specific lint image from Dockerfile
	// We run lint on the source *after* templ generation to avoid missing file errors
	_, err := dagger.Container().From("golangci/golangci-lint:v2.7.2").
		WithDirectory("/app", templSource).
		WithWorkdir("/app").
		WithExec([]string{"golangci-lint", "run", "-v", "--timeout=5m"}).
		Sync(ctx)
	if err != nil {
		return nil, fmt.Errorf("lint failed: %w", err)
	}

	// 4. Test
	// Run tests on the templ-generated source
	_, err = gobase.
		WithDirectory("/app", templSource).
		WithWorkdir("/app").
		WithExec([]string{"go", "test", "-v", "./..."}).
		Sync(ctx)
	if err != nil {
		return nil, fmt.Errorf("tests failed: %w", err)
	}

	// 5. Go Build
	// Prepare source: overlay css
	// We take the templSource and overlay the CSS generated in step 1
	buildSource := templSource.
		WithDirectory("static/css", cssDir)

	builder := gobase.
		WithDirectory("/app", buildSource).
		WithWorkdir("/app").
		WithEnvVariable("GOCACHE", "/go-build-cache").
		WithEnvVariable("GOMODCACHE", "/go-mod-cache").
		WithMountedCache("/go-build-cache", dagger.CacheVolume("go-build-cache")).
		WithMountedCache("/go-mod-cache", dagger.CacheVolume("go-mod-cache")).
		WithExec([]string{
			"go", "build",
			"-ldflags", "-X readwillbe/version.Tag=" + version,
			"-o", "/readwillbe",
			"./cmd/readwillbe/",
		})

	binary := builder.File("/readwillbe")

	// 6. Release
	// Create the final alpine image
	release := dagger.Container().From("alpine:3.20").
		WithExec([]string{"apk", "add", "--no-cache", "tzdata", "ca-certificates"}).
		WithFile("/readwillbe", binary).
		// Add nonroot user. Dockerfile does this by copying a passwd file.
		// We can achieve the same by appending to /etc/passwd.
		WithExec([]string{"sh", "-c", "echo 'nonroot:x:10001:10001:NonRoot User:/:/sbin/nologin' >> /etc/passwd"}).
		WithEnvVariable("TZ", "America/New_York").
		WithEnvVariable("PORT", ":8080").
		WithExposedPort(8080).
		WithUser("10001").
		WithEntrypoint([]string{"/readwillbe"})

	return release, nil
}
