package main

import (
	"context"
	"fmt"

	"dagger/readwillbe/internal/dagger"
)

type Readwillbe struct{}

func (m *Readwillbe) Build(
	ctx context.Context,
	source *dagger.Directory,
	// +optional
	// +default="dev"
	version string,
) (*dagger.Container, error) {
	if _, err := m.Lint(ctx, source); err != nil {
		return nil, fmt.Errorf("lint failed: %w", err)
	}

	if _, err := m.Test(ctx, source); err != nil {
		return nil, fmt.Errorf("test failed: %w", err)
	}

	templSource := m.TemplGenerate(source)
	cssDir := m.BuildCss(source)

	buildSource := templSource.WithDirectory("static/css", cssDir)

	return m.BuildBinary(buildSource, version), nil
}

func (m *Readwillbe) Lint(ctx context.Context, source *dagger.Directory) (string, error) {
	templSource := m.TemplGenerate(source)
	return dag.GolangciLint().
		WithModuleCache(dag.CacheVolume("go-mod-cache")).
		WithLinterCache(dag.CacheVolume("golangci-lint-cache")).
		Run(templSource).
		Stdout(ctx)
}

func (m *Readwillbe) Test(ctx context.Context, source *dagger.Directory) (string, error) {
	templSource := m.TemplGenerate(source)
	return dag.Container().
		From("golang:1.25-alpine").
		WithEnvVariable("GOCACHE", "/go-build-cache").
		WithEnvVariable("GOMODCACHE", "/go-mod-cache").
		WithMountedCache("/go-build-cache", dag.CacheVolume("go-build-cache")).
		WithMountedCache("/go-mod-cache", dag.CacheVolume("go-mod-cache")).
		WithDirectory("/app", templSource).
		WithWorkdir("/app").
		WithExec([]string{"go", "test", "-v", "./..."}).
		Stdout(ctx)
}

func (m *Readwillbe) BuildCss(source *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From("ghcr.io/jwhumphries/frontend:latest").
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithExec([]string{"bun", "install"}).
		WithExec([]string{"bun", "run", "build"}).
		Directory("/app/static/css")
}

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

func (m *Readwillbe) Fmt(source *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From("golang:1.25-alpine").
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithExec([]string{"go", "fmt", "./..."}).
		Directory("/app")
}
