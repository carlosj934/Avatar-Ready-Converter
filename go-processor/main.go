package main

import (
	"fmt"
	"log"
	"net/http"

	"avatar-ready-converter/processor/config"
	"avatar-ready-converter/processor/handlers"
)

func main() {
	cfg := config.Load()

	// Setup routes
	http.HandleFunc("/health", healthHandler)
	http.Handle("/process", handlers.NewProcessHandler(cfg))

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Go processor service starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
