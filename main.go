package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"time"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func customHeadersMiddleware(headers map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func noCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s]: %s\n", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}

func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}
	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}

func main() {

	port := 6969
	flag.IntVar(&port, "port", 6969, "the port the server runs on")

	bind := "localhost"
	flag.StringVar(&bind, "bind", "localhost", "the network interface to bind to (use 0.0.0.0 for all interfaces)")

	dir := ""
	flag.StringVar(&dir, "dir", "", "directory to serve (defaults to current working directory)")

	quiet := false
	flag.BoolVar(&quiet, "quiet", false, "disable logging")

	cors := false
	flag.BoolVar(&cors, "cors", false, "enable CORS headers")

	open := false
	flag.BoolVar(&open, "open", false, "automatically open browser on startup")

	customHeaders := ""
	flag.StringVar(&customHeaders, "headers", "", "custom headers in format 'Key1:Value1,Key2:Value2'")

	readTimeout := 15
	flag.IntVar(&readTimeout, "read-timeout", 15, "specify the http servers read timeout in seconds")

	writeTimeout := 15
	flag.IntVar(&writeTimeout, "write-timeout", 15, "specify the http servers write timeout in seconds")

	flag.Parse()

	host := fmt.Sprintf("%s:%d", bind, port)

	// Determine directory to serve
	serveDir := dir
	if serveDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		serveDir = cwd
	}

	fs := http.FileServer(http.Dir(serveDir))
	fs = noCacheMiddleware(fs)

	// Apply CORS middleware if enabled
	if cors {
		fs = corsMiddleware(fs)
	}

	// Apply custom headers middleware if provided
	if customHeaders != "" {
		headers := parseHeaders(customHeaders)
		fs = customHeadersMiddleware(headers)(fs)
	}

	if !quiet {
		log.Println("go-srv - a local web-server for the CWD")
		log.Printf("running on: %s in '%s'", host, serveDir)
		log.Printf("read-timeout: %ds, write-timeout: %ds", readTimeout, writeTimeout)
		if cors {
			log.Println("CORS: enabled")
		}
		if customHeaders != "" {
			log.Printf("custom headers: %s", customHeaders)
		}
		fs = loggingMiddleware(fs)
	}

	mux := http.NewServeMux()
	mux.Handle("/", fs)

	srv := &http.Server{
		Addr:              host,
		WriteTimeout:      time.Second * time.Duration(writeTimeout),
		ReadHeaderTimeout: time.Second * time.Duration(readTimeout),
		IdleTimeout:       time.Second * 60,
		Handler:           mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}
	}()

	// Open browser if requested
	if open {
		url := fmt.Sprintf("http://%s", host)
		if err := openBrowser(url); err != nil && !quiet {
			log.Printf("Failed to open browser: %v", err)
		}
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	srv.Shutdown(context.Background())
	os.Exit(0)
}
