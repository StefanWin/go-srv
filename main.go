package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s]: %s\n", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := 6969
	flag.IntVar(&port, "port", 6969, "")

	quiet := false
	flag.BoolVar(&quiet, "quiet", false, "")

	flag.Parse()

	host := fmt.Sprintf(":%d", port)

	if quiet {
		http.Handle("/", http.FileServer(http.Dir(".")))
	} else {
		log.Printf("running on: '%s'\n", host)
		http.Handle("/", loggingMiddleware(http.FileServer(http.Dir("."))))
	}

	http.ListenAndServe(host, nil)
}
