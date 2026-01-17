//go:build dev

package main

import (
	"bufio"
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
	"time"
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
	cmd     *exec.Cmd
	mu      sync.Mutex
}

func main() {
	log.SetFlags(0)

	services := []*Service{
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

	var wg sync.WaitGroup

	// Start services
	for _, svc := range services {
		wg.Add(1)
		go func(s *Service) {
			defer wg.Done()
			runService(s)
		}(svc)
	}

	// Start proxy
	server := startProxy()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n" + ColorYellow + "Shutting down gracefully..." + ColorReset)

	// Shutdown proxy first
	server.Close()

	// Send SIGTERM to all services
	for _, svc := range services {
		svc.mu.Lock()
		if svc.cmd != nil && svc.cmd.Process != nil {
			fmt.Printf("%s[%s] Sending SIGTERM...%s\n", svc.Color, svc.Name, ColorReset)
			svc.cmd.Process.Signal(syscall.SIGTERM)
		}
		svc.mu.Unlock()
	}

	// Wait for services to exit with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println(ColorGreen + "All services stopped cleanly." + ColorReset)
	case <-time.After(5 * time.Second):
		fmt.Println(ColorRed + "Timeout waiting for services, forcing shutdown..." + ColorReset)
		for _, svc := range services {
			svc.mu.Lock()
			if svc.cmd != nil && svc.cmd.Process != nil {
				svc.cmd.Process.Kill()
			}
			svc.mu.Unlock()
		}
	}
}

func runService(s *Service) {
	cmd := exec.Command(s.Command, s.Args...)

	s.mu.Lock()
	s.cmd = cmd
	s.mu.Unlock()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("%s[%s] Error creating stdout pipe: %v%s\n", s.Color, s.Name, err, ColorReset)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("%s[%s] Error creating stderr pipe: %v%s\n", s.Color, s.Name, err, ColorReset)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("%s[%s] Error starting: %v%s\n", s.Color, s.Name, err, ColorReset)
		return
	}

	var logWg sync.WaitGroup
	logWg.Add(2)
	go func() {
		defer logWg.Done()
		streamLogs(stdout, s.Name, s.Color)
	}()
	go func() {
		defer logWg.Done()
		streamLogs(stderr, s.Name, s.Color)
	}()

	err = cmd.Wait()
	logWg.Wait() // Wait for logs to flush

	if err != nil {
		// Check if it was killed by signal (expected during shutdown)
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				if status.Signaled() {
					fmt.Printf("%s[%s] Stopped%s\n", s.Color, s.Name, ColorReset)
					return
				}
			}
		}
		log.Printf("%s[%s] Exited with error: %v%s\n", s.Color, s.Name, err, ColorReset)
	}
}

func streamLogs(r io.Reader, name, color string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		fmt.Printf("%s[%s] %s%s\n", color, name, line, ColorReset)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("%s[%s] Log stream error: %v%s\n", color, name, err, ColorReset)
	}
}

func startProxy() *http.Server {
	templTarget, _ := url.Parse("http://127.0.0.1:7332")
	goTarget, _ := url.Parse("http://127.0.0.1:8080")

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(templTarget)
			r.Out.Host = r.In.Host

			path := r.In.URL.Path
			if strings.HasPrefix(path, "/static/") || path == "/serviceWorker.js" {
				r.SetURL(goTarget)
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
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

	go func() {
		fmt.Println("\n🚀 Dev Server running at http://localhost:7331")
		fmt.Println("   - Static assets -> Go Server (8080)")
		fmt.Println("   - Other requests -> Templ Proxy (7332) -> Go Server (8080)")
		fmt.Println("   - Press Ctrl+C to stop\n")

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Proxy server error: %v", err)
		}
	}()

	return server
}
