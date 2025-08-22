package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

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

func main() {

	port := 6969
	flag.IntVar(&port, "port", 6969, "the port the server runs on")

	quiet := false
	flag.BoolVar(&quiet, "quiet", false, "disable logging")

	readTimeout := 15
	flag.IntVar(&readTimeout, "read-timeout", 15, "specify the http servers read timeout in seconds")

	flag.Parse()

	host := fmt.Sprintf(":%d", port)

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir(cwd))
	fs = noCacheMiddleware(fs)

	if !quiet {
		log.Println("go-srv - a local web-server for the CWD")
		log.Printf("running on: %s in '%s'", host, cwd)
		log.Printf("read-timeout: %ds\n", readTimeout)
		fs = loggingMiddleware(fs)
	}

	mux := http.NewServeMux()
	mux.Handle("/", fs)

	srv := &http.Server{
		Addr:              host,
		WriteTimeout:      time.Second * 15,
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

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	srv.Shutdown(context.Background())
	os.Exit(0)
}
