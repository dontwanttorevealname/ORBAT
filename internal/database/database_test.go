package database

import (
    "os"
    "testing"
    "github.com/joho/godotenv"
    "orbat/internal/models"
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
    // Test weapon creation
    weaponID := "1000" // Using test data from SQL/Seeds/001_AddTestData.sql
    weapon, err := GetWeapon(weaponID)
    if err != nil {
        t.Fatalf("Failed to get test weapon: %v", err)
    }
    if weapon.Name != "Test Rifle" {
        t.Errorf("Expected weapon name 'Test Rifle', got '%s'", weapon.Name)
    }

    // Test weapon listing
    weapons, err := GetWeapons()
    if err != nil {
        t.Fatalf("Failed to list weapons: %v", err)
    }
    if len(weapons) < 1 {
        t.Error("Expected at least one weapon in test database")
    }
}

func TestGroupOperations(t *testing.T) {
    // Test group retrieval
    groupID := "1000" // Using test data from SQL/Seeds/001_AddTestData.sql
    group, err := GetGroup(groupID)
    if err != nil {
        t.Fatalf("Failed to get test group: %v", err)
    }
    if group.Name != "Test Group" {
        t.Errorf("Expected group name 'Test Group', got '%s'", group.Name)
    }

    // Test group members
    if len(group.DirectMembers) < 1 {
        t.Error("Expected at least one member in test group")
    }
}

func TestVehicleOperations(t *testing.T) {
    // Test vehicle retrieval
    vehicleID := "1000" // Using test data from SQL/Seeds/001_AddTestData.sql
    vehicle, err := GetVehicle(vehicleID)
    if err != nil {
        t.Fatalf("Failed to get test vehicle: %v", err)
    }
    if vehicle.Name != "Test Vehicle 1" {
        t.Errorf("Expected vehicle name 'Test Vehicle 1', got '%s'", vehicle.Name)
    }
}

func TestCountryOperations(t *testing.T) {
    // Test country details retrieval
    countries, err := GetCountries()
    if err != nil {
        t.Fatalf("Failed to get countries: %v", err)
    }
    if len(countries) < 1 {
        t.Error("Expected at least one country in test database")
    }
}