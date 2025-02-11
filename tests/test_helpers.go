package tests

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"

    _ "github.com/mattn/go-sqlite3"
)

var (
    testDB     *sql.DB
    testDBPath string
)

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

func setupTestEnv() error {
    var err error

    testDBDir := "../SQL/Database/test"
    if err := os.MkdirAll(testDBDir, 0755); err != nil {
        return fmt.Errorf("error creating test directory: %v", err)
    }

    testDBPath = filepath.Join(testDBDir, "test.db")
    os.Remove(testDBPath)

    testDB, err = sql.Open("sqlite3", testDBPath)
    if err != nil {
        return fmt.Errorf("error opening database: %v", err)
    }

    if err := runMigrations(); err != nil {
        return fmt.Errorf("error running migrations: %v", err)
    }

    return nil
}

func runMigrations() error {
    migrationsDir := "../SQL/Migrations"
    files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
    if err != nil {
        return fmt.Errorf("error finding migration files: %v", err)
    }

    sort.Strings(files)

    for _, file := range files {
        content, err := os.ReadFile(file)
        if err != nil {
            return fmt.Errorf("error reading migration file %s: %v", file, err)
        }

        parts := strings.Split(string(content), "-- +goose Down")
        if len(parts) != 2 {
            return fmt.Errorf("invalid migration file format: %s", file)
        }

        upSQL := strings.Split(parts[0], "-- +goose Up")[1]

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

func resetTestDB() error {
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

    return runMigrations()
}

func cleanupTestEnv() {
    if testDB != nil {
        testDB.Close()
    }
    os.Remove(testDBPath)
}