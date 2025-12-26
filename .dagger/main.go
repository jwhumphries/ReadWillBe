package main

import (
	"context"
	"fmt"

	"dagger/readwillbe/internal/dagger"
)

type Readwillbe struct{}

// Build runs the full build pipeline: templ, lint, test, css, and build.
func (m *Readwillbe) Build(
	ctx context.Context,
	source *dagger.Directory,
	// +optional
	// +default="dev"
	version string,
) (*dagger.Container, error) {
	templSource := m.TemplGenerate(source)

	if _, err := m.Lint(ctx, templSource); err != nil {
		return nil, fmt.Errorf("lint failed: %w", err)
	}

	if _, err := m.Test(ctx, templSource); err != nil {
		return nil, fmt.Errorf("test failed: %w", err)
	}

	cssDir := m.Minify(source)

	buildSource := templSource.WithDirectory("static/css", cssDir)

	return m.BuildBinary(buildSource, version), nil
}

// Lint runs golangci-lint on the source.
func (m *Readwillbe) Lint(ctx context.Context, source *dagger.Directory) (string, error) {
	return dag.Golangci().
		Run(source, dagger.GolangciRunOpts{
			Verbose: true,
			Timeout: "5m",
		}).
		Stdout(ctx)
}

// GolangciLintFix runs golangci-lint with the fix option and returns the modified source.
func (m *Readwillbe) GolangciLintFix(source *dagger.Directory) *dagger.Directory {
	return dag.Golangci().
		Run(source, dagger.GolangciRunOpts{
			Verbose: true,
			Timeout: "5m",
			Fix:     true,
		}).
		Directory()
}

// Test runs Go tests.
func (m *Readwillbe) Test(ctx context.Context, source *dagger.Directory) (string, error) {
	return dag.Go().
		Test(source, dagger.GoTestOpts{
			Verbose: true,
		}).
		Stdout(ctx)
}

// Minify builds the frontend assets using the project's frontend image.
func (m *Readwillbe) Minify(source *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From("ghcr.io/jwhumphries/frontend:latest").
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithExec([]string{"bun", "install"}).
		WithExec([]string{"bun", "run", "build"}).
		Directory("/app/static/css")
}

// TemplGenerate generates Templ Go code.
func (m *Readwillbe) TemplGenerate(source *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From("golang:1.25-alpine").
		WithEnvVariable("GOCACHE", "/go-build-cache").
		WithEnvVariable("GOMODCACHE", "/go-mod-cache").
		WithMountedCache("/go-build-cache", dag.CacheVolume("go-build-cache")).
		WithMountedCache("/go-mod-cache", dag.CacheVolume("go-mod-cache")).
		WithExec([]string{"apk", "add", "--no-cache", "git"}).
		WithExec([]string{"go", "install", "github.com/a-h/templ/cmd/templ@latest"}).
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithExec([]string{"templ", "generate"}).
		Directory("/app")
}

// BuildBinary builds the Go binary.
func (m *Readwillbe) BuildBinary(source *dagger.Directory, version string) *dagger.Container {
	return dag.Container().
		From("golang:1.25-alpine").
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithEnvVariable("GOCACHE", "/go-build-cache").
		WithEnvVariable("GOMODCACHE", "/go-mod-cache").
		WithMountedCache("/go-build-cache", dag.CacheVolume("go-build-cache")).
		WithMountedCache("/go-mod-cache", dag.CacheVolume("go-mod-cache")).
		WithExec([]string{
			"go", "build",
			"-ldflags", "-X readwillbe/version.Tag=" + version,
			"-o", "/readwillbe",
			"./cmd/readwillbe/",
		})
}

// Release packages the binary into a minimal Alpine image.
func (m *Readwillbe) Release(
	ctx context.Context,
	source *dagger.Directory,
	// +optional
	// +default="dev"
	version string,
) (*dagger.Container, error) {
	binaryContainer, err := m.Build(ctx, source, version)
	if err != nil {
		return nil, err
	}
	binary := binaryContainer.File("/readwillbe")

	return dag.Container().
		From("alpine:3.23").
		WithExec([]string{"apk", "add", "--no-cache", "tzdata", "ca-certificates"}).
		WithFile("/readwillbe", binary).
		WithExec([]string{"sh", "-c", "echo 'nonroot:x:10001:10001:NonRoot User:/:/sbin/nologin' >> /etc/passwd"}).
		WithEnvVariable("TZ", "America/New_York").
		WithEnvVariable("PORT", ":8080").
		WithExposedPort(8080).
		WithUser("10001").
		WithEntrypoint([]string{"/readwillbe"}), nil
}
