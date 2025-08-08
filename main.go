package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Simple calendar marks service with graceful server and modularized code

func main() {
	cfg := loadConfig()

	// ensure habits loaded
	loadHabits("habits.txt")

	// open DB (fresh schema with integer habit ids)
	db = mustOpenDB(cfg.DBPath)
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/habits", handleHabits)
	mux.HandleFunc("/api/marks", handleGetMarks)
	mux.HandleFunc("/api/toggle", handleToggle)
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/favicon.ico", handleFavicon)
	mux.HandleFunc("/favicon.svg", handleFavicon)

	// Wrap with middlewares: security headers -> gzip -> logs
	var handler http.Handler = mux
	handler = securityHeaders(handler)
	handler = gzipMiddleware(handler)
	handler = logRequests(handler)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	go func() {
		log.Printf("SHCalendar listening on http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.WriteTimeout)
	defer cancel()
	_ = server.Shutdown(ctx)
	log.Println("Server stopped")
}
