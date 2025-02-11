package tests

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var (
	testDB     *sql.DB
	testDBPath string
)

func TestMain(m *testing.M) {
	// Set up test environment
	setupTestEnv()

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupTestEnv()

	os.Exit(code)
}

func setupTestEnv() {
	var err error

	// Create test database directory
	testDBDir := "../SQL/Database/test"
	if err := os.MkdirAll(testDBDir, 0755); err != nil {
		log.Fatal("Error creating test database directory:", err)
	}

	// Set test database path
	testDBPath = filepath.Join(testDBDir, "test.db")

	// Remove existing test database if it exists
	os.Remove(testDBPath)

	// Create and open new test database
	testDB, err = sql.Open("sqlite3", testDBPath)
	if err != nil {
		log.Fatal("Error opening test database:", err)
	}

	// Run migrations on test database
	if err := runMigrations(); err != nil {
		log.Fatal("Error running migrations:", err)
	}
}

func runMigrations() error {
	migrationsDir := "../SQL/Migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("error finding migration files: %v", err)
	}

	// Sort migration files to ensure correct order
	sort.Strings(files)

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading migration file %s: %v", file, err)
		}

		// Split the file into up/down migrations
		parts := strings.Split(string(content), "-- +goose Down")
		if len(parts) != 2 {
			return fmt.Errorf("invalid migration file format: %s", file)
		}

		upSQL := strings.Split(parts[0], "-- +goose Up")[1]

		// Execute the up migration
		tx, err := testDB.Begin()
		if err != nil {
			return fmt.Errorf("error starting transaction: %v", err)
		}

		for _, statement := range strings.Split(upSQL, ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}

			if _, err := tx.Exec(statement); err != nil {
				tx.Rollback()
				return fmt.Errorf("error executing migration: %v", err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("error committing migration: %v", err)
		}
	}

	return nil
}

func cleanupTestEnv() {
	if testDB != nil {
		testDB.Close()
	}
	os.Remove(testDBPath)
}

// MockStorage provides a mock implementation for storage operations
type MockStorage struct {
	files map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		files: make(map[string][]byte),
	}
}

func (m *MockStorage) Store(filename string, data []byte) error {
	m.files[filename] = data
	return nil
}

func (m *MockStorage) Retrieve(filename string) ([]byte, error) {
	data, exists := m.files[filename]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	return data, nil
}