package tests

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sort"
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

    // Create test database directory using relative path
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
                return fmt.Errorf("error executing migration statement '%s': %v", statement, err)
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

func resetTestDB() error {
    // Drop all tables
    _, err := testDB.Exec(`
        DROP TABLE IF EXISTS group_members;
        DROP TABLE IF EXISTS groups;
        DROP TABLE IF EXISTS team_members;
        DROP TABLE IF EXISTS teams;
        DROP TABLE IF EXISTS members_weapons;
        DROP TABLE IF EXISTS members;
        DROP TABLE IF EXISTS weapons;
    `)
    if err != nil {
        return fmt.Errorf("error dropping tables: %v", err)
    }

    // Re-run migrations
    return runMigrations()
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

func (m *MockStorage) Delete(filename string) error {
    if _, exists := m.files[filename]; !exists {
        return fmt.Errorf("file not found: %s", filename)
    }
    delete(m.files, filename)
    return nil
}

func (m *MockStorage) List() []string {
    var files []string
    for filename := range m.files {
        files = append(files, filename)
    }
    return files
}

// Helper function to check if a table exists
func tableExists(tableName string) bool {
    var count int
    err := testDB.QueryRow(`
        SELECT COUNT(*) 
        FROM sqlite_master 
        WHERE type='table' AND name=?
    `, tableName).Scan(&count)
    
    return err == nil && count > 0
}

// Helper function to get table column names
func getTableColumns(tableName string) ([]string, error) {
    rows, err := testDB.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var columns []string
    for rows.Next() {
        var (
            cid      int
            name     string
            dtype    string
            notnull  bool
            dfltVal  interface{}
            pk       bool
        )
        if err := rows.Scan(&cid, &name, &dtype, &notnull, &dfltVal, &pk); err != nil {
            return nil, err
        }
        columns = append(columns, name)
    }
    return columns, nil
}