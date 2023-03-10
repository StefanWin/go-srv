package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	//go:embed VERSION
	appVersion  string
	buildTime   string
	buildCommit string
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s]: %s\n", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Printf("go-srv %s (%s %s)\n", appVersion, buildCommit, buildTime)

	port := 6969
	flag.IntVar(&port, "port", 6969, "")

	quiet := false
	flag.BoolVar(&quiet, "quiet", false, "")

	flag.Parse()

	host := fmt.Sprintf(":%d", port)

	handler := http.FileServer(http.Dir("."))

	if !quiet {
		log.Printf("running on: %s\n", host)
		handler = loggingMiddleware(handler)
	}

	srv := http.Server{
		Addr:         host,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handler,
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
