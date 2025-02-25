package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"orbat/internal/database"
	"orbat/internal/handlers"
	"orbat/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables but don't exit if .env is missing
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Info: .env file not found, using environment variables\n")
	}

	// Initialize database
	if err := database.Initialize(); err != nil {
		fmt.Printf("Fatal: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize storage
	if err := storage.Initialize(); err != nil {
		fmt.Printf("Fatal: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()

	// Initialize templates
	if err := handlers.Initialize("templates"); err != nil {
		fmt.Printf("Fatal: Failed to parse templates: %v\n", err)
		os.Exit(1)
	}

	// Set up routes
	http.HandleFunc("/", handlers.GroupsHandler)
	http.HandleFunc("/group/", func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an edit request
		if len(r.URL.Path) > 7 && r.URL.Path[len(r.URL.Path)-5:] == "/edit" {
			handlers.EditGroupHandler(w, r)
		} else {
			handlers.GroupDetailsHandler(w, r)
		}
	})
	http.HandleFunc("/add_group", handlers.AddGroupHandler)
	http.HandleFunc("/weapons", handlers.WeaponsHandler)
	http.HandleFunc("/weapon/", handlers.WeaponDetailsHandler)
	http.HandleFunc("/member/", handlers.MemberWeaponsHandler)
	http.HandleFunc("/vehicles", handlers.VehiclesHandler)
	http.HandleFunc("/vehicle/", handlers.VehicleDetailsHandler)
	http.HandleFunc("/countries", handlers.CountriesHandler)
	http.HandleFunc("/country/", handlers.CountryDetailsHandler)
	http.HandleFunc("/health", handlers.HealthCheckHandler)

	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create a server with timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server with improved logging
	fmt.Printf("Server starting on port %s\n", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Fatal: Server error: %v\n", err)
		os.Exit(1)
	}
} 
