package main

import (
	"fmt"
	"log"
	"net/http"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avatar-ready-converter/processor/config"
	"avatar-ready-converter/processor/handlers"
)

func main() {
	cfg := config.Load()

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/process", handlers.NewProcessHandler(cfg))

	// create server
	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr: addr,
		Handler: mux,
		ReadTimeout: 15 * time.Second,
		WriteTimeout: time.Duration(cfg.ProcessingTimeout+10) * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	// channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// start server in goroutine
	go func() {
		log.Printf("Go processor service starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// wait for interrupt signal
	<-quit
	log.Println("Shutdown signal received, starting graceful shutdown...")

	// create shutdown context with tiemout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server gracefully stopped")
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"go-processor"}`))	
}
