package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

type Service struct {
	Name    string
	Command string
	Args    []string
	Color   string
	Port    string // Optional check for readiness
}

func main() {
	log.SetFlags(0) // No timestamps in main log, we'll handle them

	// 1. Define Services
	services := []Service{
		{
			Name:    "GO",
			Command: "air",
			Color:   ColorCyan,
		},
		{
			Name:    "TEMPL",
			Command: "templ",
			Args: []string{
				"generate",
				"--watch",
				"--proxy=http://127.0.0.1:8080",
				"--proxyport=7332",
				"--proxybind=127.0.0.1",
				"--open-browser=false",
			},
			Color: ColorYellow,
		},
		{
			Name:    "JS",
			Command: "bun",
			Args:    []string{"run", "watch:js"},
			Color:   ColorGreen,
		},
		{
			Name:    "CSS",
			Command: "bun",
			Args:    []string{"run", "watch:css"},
			Color:   ColorPurple,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// 2. Start Services
	for _, svc := range services {
		wg.Add(1)
		go func(s Service) {
			defer wg.Done()
			runService(ctx, s)
		}(svc)
	}

	// 3. Start Proxy Server
	wg.Add(1)
	go func() {
		defer wg.Done()
		startProxy()
	}()

	// 4. Handle Signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-	sigChan
	fmt.Println("\nStopping development environment...")
	cancel()
	wg.Wait()
}

func runService(ctx context.Context, s Service) {
	cmd := exec.CommandContext(ctx, s.Command, s.Args...)
	
	// Create pipes for stdout/stderr
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Printf("%s[%s] Error starting: %v%s\n", s.Color, s.Name, err, ColorReset)
		return
	}

	// Stream logs
	go streamLogs(stdout, s.Name, s.Color)
	go streamLogs(stderr, s.Name, s.Color)

	// Wait for exit
	err := cmd.Wait()
	if err != nil && ctx.Err() == nil {
		// Only log error if not cancelled
		log.Printf("%s[%s] Exited with error: %v%s\n", s.Color, s.Name, err, ColorReset)
	}
}

func streamLogs(r io.Reader, name, color string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// Clean up some common noisy logs if needed
		if strings.TrimSpace(line) == "" {
			continue
		}
		fmt.Printf("%s[%s] %s%s\n", color, name, line, ColorReset)
	}
}

func startProxy() {
	// Targets
	templTarget, _ := url.Parse("http://127.0.0.1:7332")
	goTarget, _ := url.Parse("http://127.0.0.1:8080")

	// Proxy Handler
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			// By default forward to Templ
			r.SetURL(templTarget)
			r.Out.Host = r.In.Host // Keep original host header

			// Check if it's a static asset
			path := r.In.URL.Path
			if strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/serviceWorker.js") {
				// Forward directly to Go server, bypassing Templ
				r.SetURL(goTarget)
				// fmt.Printf("Proxying static asset %s to Go server\n", path)
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			// Suppress errors during startup/restart
			if !strings.Contains(err.Error(), "connection refused") {
				log.Printf("Proxy error: %v", err)
			}
			w.WriteHeader(http.StatusBadGateway)
		},
	}

	server := &http.Server{
		Addr:    ":7331",
		Handler: proxy,
	}

	fmt.Println("\n🚀 Dev Server running at http://localhost:7331")
	fmt.Println("   - Static assets -> Go Server (8080)")
	fmt.Println("   - Other requests -> Templ Proxy (7332) -> Go Server (8080)")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("Proxy server failed: %v", err)
	}
}
