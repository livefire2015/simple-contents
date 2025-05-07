package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/torpago/simple-content-service/repository/memory"
	"github.com/torpago/simple-content-service/service"
	"github.com/torpago/simple-content-service/storage/memorystorage"
	transportHttp "github.com/torpago/simple-content-service/transport/http"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	// Create repository and storage implementations
	// For this example, we'll use in-memory implementations
	repo := memory.NewMemoryRepository()
	storage := memorystorage.NewMemoryStorage()

	// Create content service
	contentService := service.NewContentService(repo, storage)

	// Create HTTP handler
	contentHandler := transportHttp.NewContentHandler(contentService)

	// Create router and register routes
	router := chi.NewRouter()
	contentHandler.RegisterRoutes(router)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: router,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Starting server on port %d", *port)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for interrupt signal to gracefully shut down the server
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)
	case <-shutdown:
		log.Println("Shutting down server...")

		// Create a deadline for server shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Attempt to gracefully shut down the server
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
			server.Close()
		}
	}
}
