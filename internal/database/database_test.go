package database

import (
    "os"
    "testing"
    "github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
    // Load test environment variables
    if err := godotenv.Load("../../.env.test"); err != nil {
        panic("Error loading .env.test file")
    }

    // Initialize database connection
    if err := Initialize(); err != nil {
        panic("Could not initialize test database: " + err.Error())
    }

    // Run tests
    code := m.Run()

    // Cleanup
    Close()
    os.Exit(code)
}

func TestWeaponOperations(t *testing.T) {
    // Test weapons retrieval
    weapons, err := GetWeapons()
    if err != nil {
        t.Fatalf("Failed to get weapons: %v", err)
    }

    // Check if test weapons exist
    found := false
    for _, w := range weapons {
        if w.Name == "Test Rifle" {
            found = true
            if w.Type != "Test Type" {
                t.Errorf("Expected weapon type 'Test Type', got '%s'", w.Type)
            }
            break
        }
    }
    if !found {
        t.Error("Test weapon not found in database")
    }
}

func TestGroupOperations(t *testing.T) {
    // Test groups retrieval
    groups, err := GetGroups()
    if err != nil {
        t.Fatalf("Failed to get groups: %v", err)
    }

    // Check if test group exists
    found := false
    for _, g := range groups {
        if g.ID == 1000 {
            found = true
            if g.Name != "Test Group" {
                t.Errorf("Expected group name 'Test Group', got '%s'", g.Name)
            }
            break
        }
    }
    if !found {
        t.Error("Test group not found in database")
    }
}

func TestCountryOperations(t *testing.T) {
    // Test country retrieval
    countries, err := GetCountries()
    if err != nil {
        t.Fatalf("Failed to get countries: %v", err)
    }

    // Check if we have at least one country
    if len(countries) == 0 {
        t.Error("No countries found in database")
    }

    // Test country details retrieval
    if len(countries) > 0 {
        details, err := GetCountryDetails(countries[0])
        if err != nil {
            t.Fatalf("Failed to get country details: %v", err)
        }

        if details.Name == "" {
            t.Error("Country details name is empty")
        }
    }
}

func TestVehicleUsage(t *testing.T) {
    // Test getting country details which includes vehicle usage
    details, err := GetCountryDetails("Test Country")
    if err != nil {
        t.Fatalf("Failed to get country details: %v", err)
    }

    // Check vehicle usage data
    for _, v := range details.Vehicles {
        if v.Name == "Test Vehicle 1" {
            if v.InstanceCount == 0 {
                t.Error("Expected non-zero instance count for test vehicle")
            }
            return
        }
    }
    t.Error("Test vehicle not found in usage data")
}

func TestWeaponUsage(t *testing.T) {
    // Test getting country details which includes weapon usage
    details, err := GetCountryDetails("Test Country")
    if err != nil {
        t.Fatalf("Failed to get country details: %v", err)
    }

    // Check weapon usage data
    for _, w := range details.Weapons {
        if w.Name == "Test Rifle" {
            if w.UserCount == 0 {
                t.Error("Expected non-zero user count for test weapon")
            }
            return
        }
    }
    t.Error("Test weapon not found in usage data")
}