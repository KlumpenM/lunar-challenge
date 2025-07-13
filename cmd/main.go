package main

import (
	"log"
	"net/http"
	"time"

	_ "lunar-backend-challenge/docs"
	"lunar-backend-challenge/internal/api"
	"lunar-backend-challenge/internal/middleware"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Create the API handler
	apiHandler := api.NewAPIHandler()

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Set up API routes
	mux.HandleFunc("POST /messages", apiHandler.HandleMessage)
	mux.HandleFunc("GET /rockets", apiHandler.HandleGetRockets)
	mux.HandleFunc("GET /rockets/{id}", apiHandler.HandleGetRocket)
	mux.HandleFunc("GET /debug/rockets", apiHandler.HandleDebugAll)
	mux.HandleFunc("GET /debug/rockets/{id}", apiHandler.HandleDebugRocket)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Apply middleware
	handler := middleware.ChainMiddleware(mux,
		middleware.ErrorHandler,
		middleware.ContentTypeJSON,
	)

	// Simple server setup
	server := &http.Server{
		Addr:         ":8088",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("Starting Lunar Rocket Tracking API on :8088")
	log.Fatal(server.ListenAndServe())
}
