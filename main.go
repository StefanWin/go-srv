package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
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
		log.Printf("[%s]: %s", r.Method, r.URL)
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
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// browserURL returns a URL a browser can open for the given bind address.
// Wildcard addresses (0.0.0.0, ::) are not reachable as a destination on
// every platform, so they are replaced with localhost.
func browserURL(bind string, port int) string {
	host := bind
	if ip := net.ParseIP(bind); bind == "" || (ip != nil && ip.IsUnspecified()) {
		host = "localhost"
	}
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port))
}

func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}
	for pair := range strings.SplitSeq(headerStr, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(pair), ":")
		if ok {
			headers[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	return headers
}

// buildHandler assembles the middleware chain. Custom headers are applied
// innermost so they run last and can override the no-cache and CORS defaults.
func buildHandler(serveDir string, headers map[string]string, cors, logging bool) http.Handler {
	var h http.Handler = http.FileServer(http.Dir(serveDir))
	if len(headers) > 0 {
		h = customHeadersMiddleware(headers)(h)
	}
	h = noCacheMiddleware(h)
	if cors {
		h = corsMiddleware(h)
	}
	if logging {
		h = loggingMiddleware(h)
	}
	return h
}

func main() {
	port := flag.Int("port", 6969, "the port the server runs on")
	bind := flag.String("bind", "localhost", "the network interface to bind to (use 0.0.0.0 for all interfaces)")
	dir := flag.String("dir", "", "directory to serve (defaults to current working directory)")
	quiet := flag.Bool("quiet", false, "disable logging")
	cors := flag.Bool("cors", false, "enable CORS headers")
	open := flag.Bool("open", false, "automatically open browser on startup")
	customHeaders := flag.String("headers", "", "custom headers in format 'Key1:Value1,Key2:Value2'")
	readTimeout := flag.Int("read-timeout", 15, "specify the http servers read timeout in seconds")
	writeTimeout := flag.Int("write-timeout", 15, "specify the http servers write timeout in seconds")
	flag.Parse()

	host := net.JoinHostPort(*bind, strconv.Itoa(*port))

	// Determine directory to serve
	serveDir := *dir
	if serveDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		serveDir = cwd
	}
	if info, err := os.Stat(serveDir); err != nil {
		log.Fatal(err)
	} else if !info.IsDir() {
		log.Fatalf("%s is not a directory", serveDir)
	}

	headers := parseHeaders(*customHeaders)

	srv := &http.Server{
		Addr:              host,
		Handler:           buildHandler(serveDir, headers, *cors, !*quiet),
		ReadTimeout:       time.Second * time.Duration(*readTimeout),
		ReadHeaderTimeout: time.Second * time.Duration(*readTimeout),
		WriteTimeout:      time.Second * time.Duration(*writeTimeout),
		IdleTimeout:       time.Second * 60,
	}

	// Listen before anything else so a busy port fails immediately and the
	// browser is only opened once the server can accept connections.
	ln, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatal(err)
	}

	if !*quiet {
		log.Println("go-srv - a local web-server for the CWD")
		log.Printf("running on: %s in '%s'", ln.Addr(), serveDir)
		log.Printf("read-timeout: %ds, write-timeout: %ds", *readTimeout, *writeTimeout)
		if *cors {
			log.Println("CORS: enabled")
		}
		if len(headers) > 0 {
			log.Printf("custom headers: %s", *customHeaders)
		}
	}

	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	if *open {
		if err := openBrowser(browserURL(*bind, *port)); err != nil && !*quiet {
			log.Printf("Failed to open browser: %v", err)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !*quiet {
		log.Printf("shutdown: %v", err)
	}
}
