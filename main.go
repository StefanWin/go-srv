package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := 6969
	flag.IntVar(&port, "port", 6969, "")
	flag.Parse()
	host := fmt.Sprintf(":%d", port)
	log.Printf("running on: '%s'\n", host)
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(host, nil)
}
