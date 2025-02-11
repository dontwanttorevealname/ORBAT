package tests

import (
    "log"
    "os"
    "testing"
)

func TestMain(m *testing.M) {
    if err := setupTestEnv(); err != nil {
        log.Fatalf("Failed to setup test environment: %v", err)
    }

    code := m.Run()

    cleanupTestEnv()
    os.Exit(code)
}