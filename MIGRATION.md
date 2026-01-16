# Migration Guide: HTMX to React with Templ

This document outlines the migration strategy for transitioning ReadWillBe from an HTMX-based architecture to a React-based "islands of interactivity" pattern while maintaining Go, Templ, and Tailwind CSS.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Project Setup](#project-setup)
3. [Package Dependencies](#package-dependencies)
4. [Build Pipeline Updates](#build-pipeline-updates)
5. [Docker Configuration Updates](#docker-configuration-updates)
6. [Creating React Components](#creating-react-components)
7. [Integrating React with Templ](#integrating-react-with-templ)
8. [Migrating HTMX Patterns](#migrating-htmx-patterns)
9. [Using goshipit (gsi) for DaisyUI](#using-goshipit-gsi-for-daisyui)
10. [CSRF Token Handling](#csrf-token-handling)
11. [Migration Priority Order](#migration-priority-order)
12. [File-by-File Migration Guide](#file-by-file-migration-guide)

---

## Architecture Overview

### Current Architecture (HTMX)
```
Browser ←→ HTMX ←→ Go/Echo Server ←→ Templ (Server-rendered HTML)
```

### Target Architecture (React Islands)
```
Browser ←→ React Islands + Templ Static ←→ Go/Echo Server
                    ↓
              Static HTML from Templ
                    +
              React for interactive components
```

### Key Principles

1. **Server-first rendering**: Templ continues to render the initial HTML
2. **React Islands**: React mounts into specific DOM elements for interactivity
3. **Progressive Enhancement**: Core functionality works without JavaScript
4. **DaisyUI Priority**: Use DaisyUI components via goshipit; React only for complex state management
5. **Shared Tailwind**: React components use the same Tailwind/DaisyUI classes

---

## Project Setup

### Directory Structure

Create the following new directories:

```
readwillbe/
├── assets/
│   └── js/                   # React source files
│       ├── index.tsx         # Main entry point with component registry
│       ├── components/       # React components
│       │   ├── PlanEditor.tsx
│       │   ├── NotificationBell.tsx
│       │   ├── ReadingList.tsx
│       │   └── ...
│       ├── hooks/            # Custom React hooks
│       │   ├── useFetch.ts
│       │   ├── usePolling.ts
│       │   └── useCsrf.ts
│       └── types/            # TypeScript types
│           └── index.ts
├── tools/
│   └── build.js              # esbuild configuration (optional, can use bun)
├── views/
│   └── react.templ           # Templ helper for mounting React components
├── internal/
│   └── views/
│       └── components/       # goshipit DaisyUI components (via gsi add)
├── static/
│   ├── js/
│   │   └── bundle.js         # Compiled React bundle (generated)
│   ├── css/
│   │   └── main.css          # Tailwind CSS (existing)
│   ├── push-setup.js         # Keep as-is (already vanilla JS)
│   └── serviceWorker.js      # Keep as-is (must be vanilla JS)
├── tsconfig.json             # TypeScript configuration (root level)
└── ...
```

### TypeScript Configuration

Create `tsconfig.json` in the project root:

```json
{
  "compilerOptions": {
    "target": "es2020",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "module": "esnext",
    "moduleResolution": "node",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "react-jsx"
  },
  "include": ["assets/js/**/*"]
}
```

---

## Package Dependencies

### Updated package.json

Update the `package.json` to include React and build tools:

```json
{
  "name": "readwillbe",
  "version": "0.1.0",
  "description": "Reading planner and tracker",
  "private": true,
  "scripts": {
    "init": "tailwindcss -i ./input.css -o ./static/css/main.css",
    "dev": "tailwindcss -i ./input.css -o ./static/css/main.css --watch",
    "build": "tailwindcss -i ./input.css -o ./static/css/main.css --minify",
    "build:js": "node tools/build.js",
    "watch:js": "node tools/build.js --watch",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "react": "^19.1.0",
    "react-dom": "^19.1.0"
  },
  "devDependencies": {
    "@tailwindcss/cli": "4.1.18",
    "@tailwindcss/typography": "0.5.19",
    "@types/react": "^19.1.8",
    "@types/react-dom": "^19.1.6",
    "daisyui": "5.5.14",
    "esbuild": "^0.27.2",
    "tailwindcss": "4.1.18",
    "typescript": "^5.8.3"
  }
}
```

### esbuild Configuration

Create `tools/build.js`:

```javascript
const esbuild = require('esbuild');

const watch = process.argv.includes('--watch');

const ctx = esbuild.context({
    entryPoints: ['assets/js/index.tsx'],
    bundle: true,
    minify: !watch,
    sourcemap: true,
    outfile: 'static/js/bundle.js',
    logLevel: 'info',
    loader: { '.tsx': 'tsx', '.ts': 'ts' },
}).then(ctx => {
    if (watch) {
        ctx.watch();
        console.log('Watching for changes...');
    } else {
        ctx.rebuild().then(() => ctx.dispose());
    }
});
```

### Installation

```bash
bun install
```

---

## Build Pipeline Updates

### Updated Dagger Pipeline

Update `.dagger/main.go` to include a combined `BuildAssets` function that builds both CSS and React in one step:

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"dagger/readwillbe/internal/dagger"
)

type Readwillbe struct{}

func (m *Readwillbe) gitVersion(ctx context.Context, git *dagger.Directory) (string, error) {
	if git == nil {
		return "dev", nil
	}
	out, err := dag.Container().
		From("alpine/git:latest").
		WithMountedDirectory("/src/.git", git).
		WithWorkdir("/src").
		WithExec([]string{"git", "describe", "--tags", "--always"}).
		Stdout(ctx)
	if err != nil {
		return "dev", nil
	}
	return strings.TrimSpace(out), nil
}

func (m *Readwillbe) Version(
	ctx context.Context,
	// +optional
	// +defaultPath="/.git"
	git *dagger.Directory,
) (string, error) {
	return m.gitVersion(ctx, git)
}

func (m *Readwillbe) Build(
	ctx context.Context,
	source *dagger.Directory,
	// +optional
	// +defaultPath="/.git"
	git *dagger.Directory,
	// +optional
	version string,
) (*dagger.Container, error) {
	if version == "" {
		v, err := m.gitVersion(ctx, git)
		if err != nil {
			return nil, fmt.Errorf("version detection failed: %w", err)
		}
		version = v
	}

	templSource := m.TemplGenerate(source)

	if _, err := m.lintSource(ctx, templSource); err != nil {
		return nil, fmt.Errorf("lint failed: %w", err)
	}

	if _, err := m.testSource(ctx, templSource); err != nil {
		return nil, fmt.Errorf("test failed: %w", err)
	}

	// Build CSS and React together
	assetsDir := m.BuildAssets(source)
	buildSource := templSource.WithDirectory("static", assetsDir)

	return m.BuildBinary(buildSource, version), nil
}

func (m *Readwillbe) Lint(ctx context.Context, source *dagger.Directory) (string, error) {
	templSource := m.TemplGenerate(source)
	return m.lintSource(ctx, templSource)
}

func (m *Readwillbe) lintSource(ctx context.Context, source *dagger.Directory) (string, error) {
	return dag.GolangciLint().
		WithModuleCache(dag.CacheVolume("go-mod-cache")).
		WithLinterCache(dag.CacheVolume("golangci-lint-cache")).
		Run(source).
		Stdout(ctx)
}

func (m *Readwillbe) Test(ctx context.Context, source *dagger.Directory) (string, error) {
	templSource := m.TemplGenerate(source)
	return m.testSource(ctx, templSource)
}

func (m *Readwillbe) testSource(ctx context.Context, source *dagger.Directory) (string, error) {
	return dag.Container().
		From("golang:1.25-alpine").
		WithEnvVariable("GOCACHE", "/go-build-cache").
		WithEnvVariable("GOMODCACHE", "/go-mod-cache").
		WithMountedCache("/go-build-cache", dag.CacheVolume("go-build-cache")).
		WithMountedCache("/go-mod-cache", dag.CacheVolume("go-mod-cache")).
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithExec([]string{"go", "test", "-v", "./..."}).
		Stdout(ctx)
}

// BuildAssets compiles both CSS (Tailwind) and React/TypeScript in one step
func (m *Readwillbe) BuildAssets(source *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From("ghcr.io/jwhumphries/frontend:latest").
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithExec([]string{"bun", "install"}).
		WithExec([]string{"bun", "run", "build"}).    // CSS
		WithExec([]string{"bun", "run", "build:js"}). // React
		Directory("/app/static")
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
	// +defaultPath="/.git"
	git *dagger.Directory,
	// +optional
	version string,
) (*dagger.Container, error) {
	binaryContainer, err := m.Build(ctx, source, git, version)
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
```

### Updated Taskfile.yml

Add these new tasks to `Taskfile.yml`:

```yaml
  build-assets:
    desc: Compile CSS and React/TypeScript
    cmds:
      - dagger -m .dagger call build-assets --source=. export --path=./static

  typecheck:
    desc: Type-check TypeScript files
    cmds:
      - bun run typecheck
```

---

## Docker Configuration Updates

### Updated docker-compose.dev.yml

The `frontend` service now runs both CSS and React watch modes:

```yaml
services:
  templ:
    image: readwillbe:dev
    working_dir: /app
    volumes:
      - .:/app
      - go-mod-cache:/go/pkg/mod
    ports:
      - "7331:7331"
    command: >
      templ generate --watch
      --proxy="http://app:8080"
      --proxyport=7331
      --proxybind="0.0.0.0"
      --open-browser=false
    depends_on:
      app:
        condition: service_started
    networks:
      - devnet

  frontend:
    image: ghcr.io/jwhumphries/frontend:latest@sha256:682cee3e8392ecaf2e6bfdf2d4f6886e95a3fdea7efe06398d924a50e9017690
    working_dir: /app
    volumes:
      - .:/app
    command: sh -c "bun install && bun run watch:js & exec bun run dev"
    stdin_open: true
    tty: true
    networks:
      - devnet

  css-reload:
    image: readwillbe:dev
    working_dir: /app
    volumes:
      - ./static:/app/static
    environment:
      - TEMPL_EXPERIMENT=rawgo
    command: >
      air
      --build.cmd "templ generate --notify-proxy"
      --build.bin "true"
      --build.include_ext "css,js"
      --build.include_dir "static/css,static/js"
      --build.exclude_dir ""
      --build.delay 100
    depends_on:
      - templ
      - frontend
    networks:
      - devnet

  app:
    image: readwillbe:dev
    working_dir: /app
    volumes:
      - .:/app
      - go-mod-cache:/go/pkg/mod
      - go-build-cache:/root/.cache/go-build
    environment:
      - READWILLBE_PORT=:8080
      - READWILLBE_DB_PATH=/app/data/readwillbe.db
      - READWILLBE_COOKIE_SECRET=${READWILLBE_COOKIE_SECRET:-dev-only-local-secret-min-32-chars}
      - READWILLBE_SEED_DB=true
      - READWILLBE_ALLOW_SIGNUP=true
      - READWILLBE_LOG_LEVEL=debug
      - TEMPL_EXPERIMENT=rawgo
      - READWILLBE_VAPID_PUBLIC_KEY=${READWILLBE_VAPID_PUBLIC_KEY}
      - READWILLBE_VAPID_PRIVATE_KEY=${READWILLBE_VAPID_PRIVATE_KEY}
      - READWILLBE_HOSTNAME=localhost:7331
    command: sh -c "go mod download && mkdir -p tmp data && air"
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "wget", "--spider", "--quiet", "http://localhost:8080/healthz"]
      interval: 10s
      timeout: 3s
      retries: 3
      start_period: 5s
    networks:
      - devnet

volumes:
  go-mod-cache:
  go-build-cache:

networks:
  devnet:
    driver: bridge
```

---

## Creating React Components

### Entry Point (assets/js/index.tsx)

Create the main entry point using a **component registry pattern**. This pattern uses `data-react-component` attributes to declaratively mount React components:

```tsx
import React from 'react';
import { createRoot } from 'react-dom/client';

// Import components
import { PlanEditor } from './components/PlanEditor';
import { NotificationBell } from './components/NotificationBell';
import { ReadingList } from './components/ReadingList';
import { ConfirmModal } from './components/ConfirmModal';

// Component registry - maps component names to React components
const components: Record<string, React.ComponentType<any>> = {
    'PlanEditor': PlanEditor,
    'NotificationBell': NotificationBell,
    'ReadingList': ReadingList,
    'ConfirmModal': ConfirmModal,
};

// Mount all React components when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    const mounts = document.querySelectorAll('[data-react-component]');
    mounts.forEach(mount => {
        const componentName = mount.getAttribute('data-react-component');
        const propsJson = mount.getAttribute('data-props');

        if (componentName && components[componentName]) {
            const Component = components[componentName];
            const props = propsJson ? JSON.parse(propsJson) : {};
            const root = createRoot(mount);
            root.render(<Component {...props} />);
        } else {
            console.warn(`React component '${componentName}' not found in registry.`);
        }
    });
});
```

**Key benefits of this pattern:**
- Declarative: Components are mounted based on HTML attributes, not imperative JavaScript calls
- Scalable: Adding new components only requires adding them to the registry
- Clean separation: Templ doesn't need to know about React internals
- Debugging: Easy to see which components are mounted by inspecting the DOM

### Types (assets/js/types/index.ts)

```typescript
export interface Reading {
  id: number;
  planId: number;
  date: string;
  content: string;
  status: 'pending' | 'completed';
  completedAt?: string;
  plan?: Plan;
}

export interface ManualReading {
  id: string;
  date: string;
  content: string;
}

export interface Plan {
  id: number;
  title: string;
  status: 'active' | 'processing' | 'failed';
  errorMessage?: string;
  readings: Reading[];
}

export interface PlanGroup {
  plan: Plan;
  readings: Reading[];
}

export interface FetchOptions extends RequestInit {
  csrfToken?: string;
}
```

### Custom Hooks

#### useCsrf Hook (assets/js/hooks/useCsrf.ts)

```typescript
export function getCsrfToken(): string {
  const match = document.cookie.match(/(?:^|; )_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : '';
}

export function useCsrf(): string {
  return getCsrfToken();
}
```

#### useFetch Hook (assets/js/hooks/useFetch.ts)

```typescript
import { useState, useCallback } from 'react';
import { getCsrfToken } from './useCsrf';

interface UseFetchOptions {
  onSuccess?: (data: unknown) => void;
  onError?: (error: Error) => void;
}

export function useFetch<T = unknown>(options: UseFetchOptions = {}) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState<T | null>(null);

  const execute = useCallback(async (url: string, init: RequestInit = {}) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(url, {
        ...init,
        headers: {
          'X-CSRF-Token': getCsrfToken(),
          ...init.headers,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      // Check if response is HTML (for full page responses)
      const contentType = response.headers.get('content-type');
      if (contentType?.includes('text/html')) {
        // For HTML responses, replace the body
        const html = await response.text();
        document.body.innerHTML = html;
        return null;
      }

      const result = await response.json();
      setData(result as T);
      options.onSuccess?.(result);
      return result as T;
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      options.onError?.(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [options]);

  return { execute, loading, error, data };
}
```

#### usePolling Hook (assets/js/hooks/usePolling.ts)

```typescript
import { useEffect, useRef, useCallback } from 'react';

interface UsePollingOptions {
  interval: number;
  enabled?: boolean;
  immediate?: boolean;
}

export function usePolling(
  callback: () => void | Promise<void>,
  options: UsePollingOptions
) {
  const { interval, enabled = true, immediate = true } = options;
  const savedCallback = useRef(callback);
  const intervalRef = useRef<number | null>(null);

  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  const start = useCallback(() => {
    if (intervalRef.current) return;

    if (immediate) {
      savedCallback.current();
    }

    intervalRef.current = window.setInterval(() => {
      savedCallback.current();
    }, interval);
  }, [interval, immediate]);

  const stop = useCallback(() => {
    if (intervalRef.current) {
      window.clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }, []);

  useEffect(() => {
    if (enabled) {
      start();
    } else {
      stop();
    }
    return stop;
  }, [enabled, start, stop]);

  return { start, stop };
}
```

---

## Integrating React with Templ

### The `@React()` Helper Pattern

The cleanest way to integrate React with Templ is using a centralized helper component. This pattern creates a `<div>` with `data-react-component` and `data-props` attributes that the React registry automatically mounts to.

#### Create views/react.templ

```templ
package views

import "encoding/json"

// React renders a mount point for a React component.
// The component will be automatically mounted by the React registry on DOMContentLoaded.
templ React(componentName string, props interface{}) {
	<div
		data-react-component={ componentName }
		data-props={ jsonProps(props) }
	></div>
}

// jsonProps converts props to JSON, handling nil as empty object
func jsonProps(props interface{}) string {
	b, _ := json.Marshal(props)
	if string(b) == "null" {
		return "{}"
	}
	return string(b)
}
```

**Key benefits:**
- **Clean**: No inline JavaScript in templ files
- **Type-safe**: Props are passed as Go structs/maps
- **Declarative**: Component mounting is automatic via the registry
- **Fallback-friendly**: Can include server-rendered fallback content inside the div

#### Updated layout.templ

Remove the HTMX script and add React bundle:

```templ
templ Layout(cfg types.Config, user *types.User, title string) {
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="UTF-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1"/>
			<title>{ title }</title>
			<link href={ versionedPath("/static/css/main.css") } rel="stylesheet"/>
			// REMOVE: <script src="https://unpkg.com/htmx.org@2.0.6"></script>
			<script type="module" src="https://unpkg.com/cally"></script>
			<script src={ versionedPath("/static/push-setup.js") }></script>
			// ADD: React bundle
			<script src={ versionedPath("/static/js/bundle.js") } defer></script>
		</head>
		// ... rest of layout
	</html>
}
```

#### Example: Using @React() for Notification Bell

**Templ Component (views/notifications.templ):**

```templ
templ NotificationBellIsland(count int) {
	@React("NotificationBell", map[string]interface{}{
		"initialCount": count,
	})
}
```

That's it. The component registry automatically mounts `NotificationBell` with the provided props.

#### Example: Using @React() for Plan Editor

**Templ Component (views/plans.templ):**

```templ
templ ManualPlanCreate(cfg types.Config, user *types.User, title string, readings []ManualReading) {
	@Layout(cfg, user, "Create Plan Manually - ReadWillBe") {
		<div class="max-w-4xl mx-auto">
			<div class="flex items-center gap-4 mb-6">
				<a href="/plans" class="btn btn-ghost btn-sm gap-2">
					@BackIcon()
					Back to Plans
				</a>
				<h1 class="text-3xl font-bold">Create Reading Plan</h1>
			</div>
			<form method="POST" action="/plans/create-manual">
				@React("PlanEditor", map[string]interface{}{
					"initialTitle":    title,
					"initialReadings": readings,
				})
			</form>
		</div>
	}
}
```

#### Example: With Server-Rendered Fallback

For progressive enhancement, include fallback content:

```templ
templ NotificationBellWithFallback(count int) {
	<div
		data-react-component="NotificationBell"
		data-props={ jsonProps(map[string]interface{}{"initialCount": count}) }
	>
		// Fallback content shown before React loads
		<button type="button" class="btn btn-ghost btn-circle" aria-label="Notifications">
			<div class="indicator">
				@BellIcon("h-5 w-5")
				if count > 0 {
					<span class="badge badge-sm badge-primary indicator-item">{ fmt.Sprintf("%d", count) }</span>
				}
			</div>
		</button>
	</div>
}
```

### Incremental Migration: Hidden Form Input Pattern

For complex forms, you can use a **hidden input pattern** that allows React to manage state while using traditional form submission. This is useful during incremental migration because it doesn't require changing backend endpoints.

**React Component (assets/js/components/PlanEditor.tsx):**

```tsx
import React, { useState } from 'react';

interface Reading {
    id: string;
    date: string;
    content: string;
}

interface PlanEditorProps {
    initialReadings?: Reading[];
    initialTitle?: string;
}

export const PlanEditor: React.FC<PlanEditorProps> = ({
    initialReadings = [],
    initialTitle = ''
}) => {
    const [title, setTitle] = useState(initialTitle);
    const [readings, setReadings] = useState<Reading[]>(
        initialReadings.map(r => ({...r, id: r.id || crypto.randomUUID()}))
    );

    const addReading = () => {
        setReadings([...readings, { id: crypto.randomUUID(), date: '', content: '' }]);
    };

    const removeReading = (id: string) => {
        setReadings(readings.filter(r => r.id !== id));
    };

    const updateReading = (id: string, field: keyof Reading, value: string) => {
        setReadings(readings.map(r => r.id === id ? { ...r, [field]: value } : r));
    };

    return (
        <div className="card bg-base-200 shadow-xl">
            <div className="card-body">
                {/* Title input */}
                <div className="form-control w-full mb-6">
                    <label className="label">
                        <span className="label-text font-bold">Plan Title</span>
                    </label>
                    <input
                        type="text"
                        name="title"
                        value={title}
                        onChange={(e) => setTitle(e.target.value)}
                        placeholder="Enter plan title"
                        className="input input-bordered w-full"
                        required
                    />
                </div>

                {/* Readings table */}
                <div className="overflow-x-auto">
                    <table className="table w-full">
                        <thead>
                            <tr>
                                <th className="w-48">Date</th>
                                <th>Reading Content</th>
                                <th className="w-16"></th>
                            </tr>
                        </thead>
                        <tbody>
                            {readings.map((reading) => (
                                <tr key={reading.id}>
                                    <td>
                                        <input
                                            type="date"
                                            value={reading.date}
                                            onChange={(e) => updateReading(reading.id, 'date', e.target.value)}
                                            className="input input-bordered input-sm w-full"
                                            required
                                        />
                                    </td>
                                    <td>
                                        <input
                                            type="text"
                                            value={reading.content}
                                            onChange={(e) => updateReading(reading.id, 'content', e.target.value)}
                                            placeholder="What to read..."
                                            className="input input-bordered input-sm w-full"
                                            required
                                        />
                                    </td>
                                    <td>
                                        <button
                                            type="button"
                                            onClick={() => removeReading(reading.id)}
                                            className="btn btn-ghost btn-xs text-error"
                                        >
                                            ✕
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>

                <button type="button" onClick={addReading} className="btn btn-outline btn-primary btn-sm">
                    + Add Reading
                </button>

                {/* HIDDEN INPUT: Serializes React state for form submission */}
                <input type="hidden" name="readingsJSON" value={JSON.stringify(readings)} />

                <div className="card-actions justify-end mt-6">
                    <a href="/plans" className="btn btn-ghost">Cancel</a>
                    <button
                        type="submit"
                        className="btn btn-primary"
                        disabled={readings.length === 0 || !title}
                    >
                        Create Plan
                    </button>
                </div>
            </div>
        </div>
    );
};
```

**Backend Handler (cmd/readwillbe/plans.go):**

```go
func (s *Server) createManualPlan(c echo.Context) error {
    title := c.FormValue("title")
    readingsJSON := c.FormValue("readingsJSON")

    var readings []views.ManualReading
    if err := json.Unmarshal([]byte(readingsJSON), &readings); err != nil {
        return c.String(http.StatusBadRequest, "Invalid readings data")
    }

    // Process readings...
}
```

**Key benefits:**
- No JSON API needed initially
- Works with standard form POST
- Graceful degradation if JS fails
- Easy to migrate to fetch-based approach later

---

## Migrating HTMX Patterns

### Pattern Reference Table

| HTMX Pattern | React Equivalent |
|--------------|------------------|
| `hx-get="/url"` | `fetch('/url')` in useEffect or event handler |
| `hx-post="/url"` | `fetch('/url', { method: 'POST', body: ... })` |
| `hx-delete="/url"` | `fetch('/url', { method: 'DELETE' })` |
| `hx-put="/url"` | `fetch('/url', { method: 'PUT', body: ... })` |
| `hx-target="#id"` | Update React state, or `document.getElementById('id').innerHTML = ...` |
| `hx-swap="outerHTML"` | Replace element via state update or DOM manipulation |
| `hx-swap="innerHTML"` | Update children via state |
| `hx-trigger="load"` | `useEffect(() => { ... }, [])` |
| `hx-trigger="every 5s"` | `usePolling` hook with interval |
| `hx-trigger="click"` | `onClick` event handler |
| `hx-disabled-elt="find button"` | `disabled={loading}` prop |
| `hx-on::after-request` | Promise `.then()` or `onSuccess` callback |

### Example Migrations

#### 1. Form Submission (Auth)

**Before (HTMX):**
```templ
<form hx-post="/auth/sign-in" hx-target="body" hx-disabled-elt="find button" class="group">
	<input type="hidden" name="_csrf" value={ csrf }/>
	<input type="email" name="email" required/>
	<input type="password" name="password" required/>
	<button type="submit" class="btn btn-primary">
		<span class="loading loading-spinner loading-sm hidden group-[.htmx-request]:inline-block"></span>
		Sign In
	</button>
</form>
```

**After (React):**
```tsx
function SignInForm({ csrf }: { csrf: string }) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    const formData = new FormData(e.currentTarget);

    try {
      const response = await fetch('/auth/sign-in', {
        method: 'POST',
        body: formData,
        headers: {
          'X-CSRF-Token': getCsrfToken(),
        },
      });

      if (response.ok) {
        // Redirect on success
        window.location.href = '/dashboard';
      } else if (response.status === 422) {
        // Handle validation error
        const html = await response.text();
        document.body.innerHTML = html;
      }
    } catch (err) {
      setError('Network error. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input type="hidden" name="_csrf" value={csrf} />
      <input type="email" name="email" required className="input input-bordered w-full" />
      <input type="password" name="password" required className="input input-bordered w-full" />
      {error && (
        <div role="alert" className="alert alert-error">
          <span>{error}</span>
        </div>
      )}
      <button type="submit" className="btn btn-primary w-full" disabled={loading}>
        {loading && <span className="loading loading-spinner loading-sm" />}
        Sign In
      </button>
    </form>
  );
}
```

#### 2. Polling (Plan Processing Status)

**Before (HTMX):**
```templ
<div hx-get="/plans" hx-trigger="every 5s" hx-select="#plans-list" hx-target="#plans-list" hx-swap="outerHTML">
	<progress class="progress progress-primary w-full"></progress>
</div>
```

**After (React):**
```tsx
function PlanProcessingStatus({ planId }: { planId: number }) {
  const [status, setStatus] = useState<'processing' | 'active' | 'failed'>('processing');

  usePolling(async () => {
    try {
      const response = await fetch(`/api/plans/${planId}/status`);
      const data = await response.json();
      setStatus(data.status);

      if (data.status !== 'processing') {
        // Refresh the page when done
        window.location.reload();
      }
    } catch (e) {
      console.error('Failed to fetch plan status', e);
    }
  }, { interval: 5000, enabled: status === 'processing' });

  if (status === 'processing') {
    return (
      <div className="mt-4">
        <progress className="progress progress-primary w-full" />
        <div className="text-sm opacity-70 mt-1 text-center">Importing readings...</div>
      </div>
    );
  }

  return null;
}
```

#### 3. Complete Reading (Dashboard Action)

**Before (HTMX):**
```templ
<form hx-post={ fmt.Sprintf("/reading/%d/complete", reading.ID) } hx-target="body" class="group" hx-disabled-elt="find button">
	<button type="submit" class="btn btn-primary btn-sm">
		<span class="loading loading-spinner loading-xs hidden group-[.htmx-request]:inline-block"></span>
		@CheckIcon("h-5 w-5 group-[.htmx-request]:hidden")
		Complete
	</button>
</form>
```

**After (React):**
```tsx
function CompleteReadingButton({ readingId, onComplete }: { readingId: number; onComplete?: () => void }) {
  const [loading, setLoading] = useState(false);

  const handleComplete = async () => {
    setLoading(true);
    try {
      const response = await fetch(`/reading/${readingId}/complete`, {
        method: 'POST',
        headers: {
          'X-CSRF-Token': getCsrfToken(),
        },
      });

      if (response.ok) {
        onComplete?.();
        window.location.reload();
      }
    } catch (e) {
      console.error('Failed to complete reading', e);
    } finally {
      setLoading(false);
    }
  };

  return (
    <button
      type="button"
      className="btn btn-primary btn-sm"
      onClick={handleComplete}
      disabled={loading}
    >
      {loading ? (
        <span className="loading loading-spinner loading-xs" />
      ) : (
        <CheckIcon className="h-5 w-5" />
      )}
      Complete
    </button>
  );
}
```

#### 4. Delete with Confirmation Modal

**Before (HTMX in components.templ):**
```templ
templ ConfirmModal(id string, title string, message string, confirmText string, hxMethod string, hxUrl string, hxTarget string) {
	<dialog id={ id } class="modal">
		<div class="modal-box">
			<h3 class="font-bold text-lg">{ title }</h3>
			<p class="py-4">{ message }</p>
			<div class="modal-action">
				<form method="dialog">
					<button class="btn">Cancel</button>
				</form>
				if hxMethod == "DELETE" {
					<button hx-delete={ hxUrl } hx-target={ hxTarget } class="btn btn-error">
						{ confirmText }
					</button>
				}
			</div>
		</div>
	</dialog>
}
```

**After (React):**
```tsx
interface ConfirmModalProps {
  id: string;
  title: string;
  message: string;
  confirmText: string;
  method: 'DELETE' | 'POST';
  url: string;
  onSuccess?: () => void;
}

export function ConfirmModal({ id, title, message, confirmText, method, url, onSuccess }: ConfirmModalProps) {
  const [loading, setLoading] = useState(false);
  const dialogRef = useRef<HTMLDialogElement>(null);

  const handleConfirm = async () => {
    setLoading(true);
    try {
      const response = await fetch(url, {
        method,
        headers: {
          'X-CSRF-Token': getCsrfToken(),
        },
      });

      if (response.ok) {
        dialogRef.current?.close();
        onSuccess?.();
      }
    } catch (e) {
      console.error('Action failed', e);
    } finally {
      setLoading(false);
    }
  };

  // Expose open method
  useEffect(() => {
    const element = document.getElementById(id);
    if (element) {
      (element as any).showModal = () => dialogRef.current?.showModal();
    }
  }, [id]);

  return (
    <dialog ref={dialogRef} id={id} className="modal">
      <div className="modal-box">
        <h3 className="font-bold text-lg">{title}</h3>
        <p className="py-4">{message}</p>
        <div className="modal-action">
          <form method="dialog">
            <button className="btn">Cancel</button>
          </form>
          <button
            className={`btn ${method === 'DELETE' ? 'btn-error' : 'btn-primary'}`}
            onClick={handleConfirm}
            disabled={loading}
          >
            {loading && <span className="loading loading-spinner loading-sm" />}
            {confirmText}
          </button>
        </div>
      </div>
      <form method="dialog" className="modal-backdrop">
        <button>close</button>
      </form>
    </dialog>
  );
}
```

---

## Using goshipit (gsi) for DaisyUI

### Installation

```bash
go install github.com/haatos/goshipit/cmd/gsi@latest
```

### Adding Components

```bash
# List available components
gsi list

# Add specific components
gsi add button
gsi add card
gsi add modal
gsi add dropdown
gsi add alert
gsi add badge
gsi add progress
gsi add toggle
```

Components are added to `internal/views/components/` as `.templ` files.

### When to Use goshipit vs React

| Use goshipit (DaisyUI/Templ) | Use React |
|------------------------------|-----------|
| Static display components | Complex state management |
| Simple forms without validation | Forms with real-time validation |
| Navigation menus | Interactive lists with CRUD |
| Cards, badges, alerts | Polling/live updates |
| Modals (simple) | Modals with async actions |
| Progress indicators | Real-time progress tracking |

### Example: Using gsi Button in Templ

After running `gsi add button`, you get a `button.templ` file you can use:

```templ
templ SaveButton() {
	@Button(ButtonProps{
		Class:    "btn-primary",
		Type:     "submit",
		Children: templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			_, err := io.WriteString(w, "Save Changes")
			return err
		}),
	})
}
```

---

## CSRF Token Handling

### React Hook for CSRF

All API calls must include the CSRF token. Use the centralized hook:

```typescript
// react/hooks/useCsrf.ts
export function getCsrfToken(): string {
  const match = document.cookie.match(/(?:^|; )_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : '';
}
```

### Usage in Fetch Calls

```typescript
const response = await fetch('/api/endpoint', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': getCsrfToken(),
  },
  body: JSON.stringify(data),
});
```

### For Form Submissions

```tsx
const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
  e.preventDefault();
  const formData = new FormData(e.currentTarget);

  await fetch('/endpoint', {
    method: 'POST',
    body: formData,
    headers: {
      'X-CSRF-Token': getCsrfToken(),
    },
  });
};
```

---

## Migration Priority Order

### Phase 1: Setup & Infrastructure
1. Create `react/` directory structure
2. Update `package.json` with React dependencies
3. Create `tsconfig.json`
4. Update Dagger pipeline with `BuildReact` function
5. Update `docker-compose.dev.yml` for React dev mode
6. Create entry point `react/index.tsx`
7. Create shared hooks (`useCsrf`, `useFetch`, `usePolling`)

### Phase 2: Simple Components (Low Risk)
1. **Auth forms** (`auth.templ`) - Simple form submissions
2. **Confirm modals** (`components.templ`) - Reusable confirmation dialogs
3. **Alert component** (`components.templ`) - Already stateless

### Phase 3: Medium Complexity
1. **Notification bell** (`notifications.templ`) - Polling pattern
2. **Account settings** (`account.templ`) - Forms with toggles
3. **History page** (`history.templ`) - List with uncomplete action

### Phase 4: Complex Components
1. **Dashboard** (`dashboard.templ`) - Multiple readings, complete actions
2. **Plans list** (`plans.templ`) - Processing status polling
3. **Manual plan form** (`plans.templ`) - Most complex: CRUD operations on readings

### Phase 5: Cleanup
1. Remove HTMX script from `layout.templ`
2. Remove HTMX-specific CSS classes (`.group-[.htmx-request]:*`)
3. Update loading state patterns to use React state
4. Remove inline HTMX event handlers

---

## File-by-File Migration Guide

### views/layout.templ

**Changes Required:**
1. Remove HTMX script tag
2. Add React bundle script
3. Convert notification bell to React island
4. Keep sign-out button (can use simple form POST or React)
5. Keep inline utility functions (togglePasswordVisibility, etc.)

**HTMX References to Remove:**
- Line 22: `<script src="https://unpkg.com/htmx.org@2.0.6"></script>`
- Lines 43-46: `hx-get="/notifications/count"` polling
- Line 99: `hx-post="/auth/sign-out"`
- Lines 137-149: HTMX event handlers (htmx:configRequest, htmx:beforeSwap)

### views/auth.templ

**Changes Required:**
1. Convert sign-in form to use fetch API
2. Convert sign-up form to use fetch API
3. Handle validation errors via React state
4. Navigation between sign-in/sign-up via React router or simple fetch

**HTMX References to Remove:**
- Line 15: `hx-post="/auth/sign-in" hx-target="body" hx-disabled-elt="find button"`
- Line 52: `hx-get="/auth/sign-up" hx-target="body"`
- Line 73: `hx-post="/auth/sign-up" hx-target="body" hx-disabled-elt="find button"`
- Line 113: `hx-get="/auth/sign-in" hx-target="body"`

### views/dashboard.templ

**Changes Required:**
1. Convert complete button to React component
2. Keep static card rendering via templ
3. Add React island for interactive reading rows

**HTMX References to Remove:**
- Line 74: `hx-post={ fmt.Sprintf("/reading/%d/complete", reading.ID) } hx-target="body" ... hx-disabled-elt="find button"`

### views/history.templ

**Changes Required:**
1. Convert uncomplete button to React component

**HTMX References to Remove:**
- Line 54: `hx-post={ fmt.Sprintf("/reading/%d/uncomplete", reading.ID) } hx-target="body" hx-disabled-elt="find button"`

### views/plans.templ

**Changes Required (Most Complex):**
1. Convert plan card with processing status to React
2. Convert manual plan form to full React island
3. Convert reading row CRUD operations
4. Handle polling for processing plans

**HTMX References to Remove:**
- Line 134: `hx-get="/plans" hx-trigger="every 5s"` (polling)
- Line 206: `hx-post="/plans/create" hx-target="body" hx-disabled-elt="find button"`
- Line 257: `hx-post={ fmt.Sprintf("/plans/%d/edit", plan.ID) } hx-target="body" hx-disabled-elt="find button"`
- Line 381: `hx-post="/plans/draft/title" hx-trigger="change from:#plan-title" hx-swap="none"`
- Line 469: `hx-post="/plans/draft/reading" hx-target="#manual-plan-form" hx-swap="outerHTML" hx-on::after-request="..."`
- Line 482: `hx-delete="/plans/draft" hx-swap="none"`
- Lines 485-487: `hx-post="/plans/create-manual" hx-target="body" hx-disabled-elt="this"`
- Lines 517-519: `hx-get={ "/plans/draft/reading/" + reading.ID + "/edit" }...`
- Lines 528-530: `hx-delete={ "/plans/draft/reading/" + reading.ID }...`
- Line 594: `hx-put={ "/plans/draft/reading/" + reading.ID }...`
- Lines 607-609: `hx-get={ "/plans/draft/reading/" + reading.ID }...`

### views/account.templ

**Changes Required:**
1. Convert notification settings form
2. Convert email settings form
3. Convert test email form
4. Keep push notification buttons (already vanilla JS)

**HTMX References to Remove:**
- Line 31: `hx-post="/account/settings" hx-target="body" hx-disabled-elt="find button"`
- Line 112: `hx-post="/account/settings" hx-target="body" hx-disabled-elt="find button"`
- Line 152: `hx-post="/account/test-email" hx-swap="innerHTML" hx-target="#test-email-result"`

### views/notifications.templ

**Changes Required:**
1. Convert entire notification bell to React island
2. Handle dropdown state in React
3. Implement polling for count updates

**HTMX References to Remove:**
- Lines 13-16: `hx-get="/notifications/dropdown" hx-trigger="click" hx-target="#notification-dropdown" hx-swap="innerHTML"`

### views/components.templ

**Changes Required:**
1. Convert ConfirmModal to React component
2. Keep Alert as templ (stateless)

**HTMX References to Remove:**
- Lines 26-34: `hx-delete={ hxUrl } hx-target={ hxTarget }` in modal buttons
- Lines 35-43: `hx-post={ hxUrl } hx-target={ hxTarget }` in modal buttons

---

## API Endpoint Updates

For React to work efficiently, consider adding JSON API endpoints alongside existing HTML endpoints:

### New JSON Endpoints

| Endpoint | Method | Response | Purpose |
|----------|--------|----------|---------|
| `/api/notifications/count` | GET | `{ count: number }` | Notification count for polling |
| `/api/notifications/readings` | GET | `{ readings: Reading[] }` | Dropdown content |
| `/api/plans/:id/status` | GET | `{ status: string }` | Plan processing status |
| `/api/reading/:id/complete` | POST | `{ success: boolean }` | Complete reading (JSON) |
| `/api/reading/:id/uncomplete` | POST | `{ success: boolean }` | Uncomplete reading (JSON) |

### Hybrid Approach

Keep existing HTML endpoints for full-page loads; add JSON endpoints for React islands:

```go
// Existing (HTML response)
e.POST("/reading/:id/complete", handleCompleteReading)

// New (JSON response)
e.POST("/api/reading/:id/complete", handleCompleteReadingJSON)
```

---

## Testing Strategy

### Unit Tests (React)
```bash
# Add test dependencies
bun add -d @testing-library/react @testing-library/jest-dom vitest
```

### Integration Tests
- Test React islands mounting correctly
- Test API calls with mocked fetch
- Test CSRF token inclusion

### E2E Tests
- Keep existing E2E tests
- Update selectors if needed (data-testid attributes)

---

## Rollback Plan

If issues arise during migration:

1. **Keep old templ files**: Create backups with `.htmx.templ` suffix
2. **Feature flags**: Use environment variable to toggle between HTMX/React
3. **Gradual rollout**: Migrate one page at a time, test thoroughly
4. **HTMX fallback**: Keep HTMX script as fallback for unmigrated components

```templ
templ Layout(cfg types.Config, user *types.User, title string) {
	// ...
	if cfg.UseReact {
		<script src={ versionedPath("/static/js/index.js") } defer></script>
	} else {
		<script src="https://unpkg.com/htmx.org@2.0.6"></script>
	}
	// ...
}
```

---

## Checklist

### Setup
- [ ] Create `assets/js/` directory structure
- [ ] Create `tools/build.js` for esbuild
- [ ] Update `package.json` with React dependencies and scripts
- [ ] Create `tsconfig.json` in project root
- [ ] Update `.dagger/main.go` with `BuildAssets` function
- [ ] Update `docker-compose.dev.yml` for React watch mode
- [ ] Test dev environment builds correctly (`task dev-start`)

### Core Infrastructure
- [ ] Create `views/react.templ` helper component
- [ ] Create `assets/js/index.tsx` with component registry
- [ ] Create `assets/js/hooks/useCsrf.ts`
- [ ] Create `assets/js/hooks/useFetch.ts`
- [ ] Create `assets/js/hooks/usePolling.ts`
- [ ] Create `assets/js/types/index.ts`

### goshipit DaisyUI Components
- [ ] Run `go install github.com/haatos/goshipit/cmd/gsi@latest`
- [ ] Run `gsi add button input card alert badge modal toggle`
- [ ] Review components in `internal/views/components/`

### Component Migration
- [ ] PlanEditor (with hidden form input pattern)
- [ ] NotificationBell (with polling)
- [ ] ReadingList (complete/uncomplete buttons)
- [ ] ConfirmModal
- [ ] Auth forms (SignIn, SignUp) - optional, can use plain forms
- [ ] AccountSettings forms - optional

### Cleanup
- [ ] Remove HTMX script from layout.templ
- [ ] Remove HTMX CSS patterns (`.group-[.htmx-request]:*`)
- [ ] Remove HTMX event handlers from layout.templ
- [ ] Update documentation
- [ ] Run full test suite

---

## References

- [Templ Guide - Using React](https://templ.guide/syntax-and-usage/using-react-with-templ/)
- [Templ Guide - Script Templates](https://templ.guide/syntax-and-usage/script-templates/)
- [Templ React Integration Example](https://github.com/a-h/templ/tree/main/examples/integration-react)
- [goshipit (gsi)](https://github.com/haatos/goshipit)
- [DaisyUI Components](https://daisyui.com/components/)
- [React Documentation](https://react.dev/)
- [Bun Bundler](https://bun.sh/docs/bundler)
