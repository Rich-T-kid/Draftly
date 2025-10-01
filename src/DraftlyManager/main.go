package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"Draftly/CRUD/handlers"
	"Draftly/CRUD/services"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize database service
	dbService, err := services.NewDatabaseService()
	if err != nil {
		log.Fatalf("Failed to initialize database service: %v", err)
	}
	defer dbService.Close()

	// Initialize S3 service with better error handling
	fmt.Println("DEBUG: Attempting to initialize S3 service...")
	s3Service, err := services.NewS3Service()
	if err != nil {
		log.Printf("WARNING: S3 service failed to initialize: %v", err)
		log.Printf("Continuing without S3 service - content upload will fail")
		// Don't exit - continue without S3 for now
	} else {
		fmt.Println("DEBUG: S3 service initialized successfully")
	}

	// Initialize handlers
	userHandler := handlers.NewUserHandler(dbService)
	documentHandler := handlers.NewDocumentHandler(dbService, s3Service)

	// Create router
	r := mux.NewRouter()

	// Add request logging middleware
	r.Use(loggingMiddleware)

	// API version prefix
	api := r.PathPrefix("/v1").Subrouter()

	// User routes
	api.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Document routes
	api.HandleFunc("/documents/{userId}", documentHandler.CreateDocument).Methods("POST")
	fmt.Println("DEBUG: Registered POST /documents/{userId} route")
	api.HandleFunc("/documents/{userId}", documentHandler.GetUserDocuments).Methods("GET")
	api.HandleFunc("/documents/{userId}/{documentId}", documentHandler.GetDocument).Methods("GET")
	api.HandleFunc("/documents/{userId}/{documentId}", documentHandler.UpdateDocument).Methods("PUT")
	api.HandleFunc("/documents/{userId}/{documentId}", documentHandler.DeleteDocument).Methods("DELETE")

	// Document content route (S3 update)
	api.HandleFunc("/documents/{documentId}", documentHandler.UpdateDocumentContent).Methods("PUT")

	// CORS middleware
	api.Use(corsMiddleware)

	// Health check endpoint
	r.HandleFunc("/health", healthCheck).Methods("GET")

	// Get port from environment or default to 6060
	port := os.Getenv("CRUD_PORT")
	if port == "" {
		port = "6060"
	}

	log.Printf("Starting Draftly API server on port %s", port)
	log.Printf("Health check available at: http://localhost:%s/health", port)
	log.Printf("API endpoints available at: http://localhost:%s/v1", port)

	// Start server
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request: %s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware handles CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// healthCheck provides a simple health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "draftly-api"}`))
}
