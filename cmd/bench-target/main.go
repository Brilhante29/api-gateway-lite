package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Received-Correlation-ID", r.Header.Get("X-Correlation-ID"))
		w.Header().Set("X-Received-Traceparent", r.Header.Get("traceparent"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	log.Printf("bench target listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
