package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// DB is the global database connection
var DB *sql.DB

// Initialize sets up the database connection
func Initialize() error {
	var err error
	maxRetries := 5
	
	for i := 0; i < maxRetries; i++ {
		DB, err = sql.Open("libsql", os.Getenv("DATABASE_URL"))
		if err == nil {
			// Test the connection
			if err = DB.Ping(); err == nil {
				fmt.Printf("Successfully connected to database\n")
				return nil
			}
		}
		fmt.Printf("Attempt %d: Failed to connect to database: %v\n", i+1, err)
		if i < maxRetries-1 {
			time.Sleep(time.Second * 2)
		}
	}
	
	return fmt.Errorf("could not establish database connection after %d attempts: %v", maxRetries, err)
}

// Close closes the database connection
func Close() {
	if DB != nil {
		DB.Close()
	}
} 