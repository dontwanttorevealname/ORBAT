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
    weapons, err := GetWeapons()
    if err != nil {
        t.Fatalf("Failed to get weapons: %v", err)
    }

    // Check if test weapons exist
    found := false
    for _, w := range weapons {
        if w.ID == 1000 {
            found = true
            if w.Name != "Test Rifle" {
                t.Errorf("Expected weapon name 'Test Rifle', got '%s'", w.Name)
            }
            if w.Type != "Test Type" {
                t.Errorf("Expected weapon type 'Test Type', got '%s'", w.Type)
            }
            if w.Caliber != "Test Caliber" {
                t.Errorf("Expected caliber 'Test Caliber', got '%s'", w.Caliber)
            }
            break
        }
    }
    if !found {
        t.Error("Test weapon (ID: 1000) not found in database")
    }
}

func TestGroupOperations(t *testing.T) {
    groups, err := GetGroups()
    if err != nil {
        t.Fatalf("Failed to get groups: %v", err)
    }

    // Check if test group exists
    found := false
    for _, g := range groups {
        if g.ID == 1000 {
            found = true
            if g.Size != 3 { // Should have 3 test members
                t.Errorf("Expected group size 3, got %d", g.Size)
            }
            break
        }
    }
    if !found {
        t.Error("Test group (ID: 1000) not found in database")
    }
}

func TestCountryOperations(t *testing.T) {
    countries, err := GetCountries()
    if err != nil {
        t.Fatalf("Failed to get countries: %v", err)
    }

    // Check for test country
    testCountryFound := false
    for _, country := range countries {
        if country == "Test Country" {
            testCountryFound = true
            break
        }
    }
    if !testCountryFound {
        t.Error("Test Country not found in countries list")
    }

    // Test country details
    details, err := GetCountryDetails("Test Country")
    if err != nil {
        t.Fatalf("Failed to get country details: %v", err)
    }

    if details.Name != "Test Country" {
        t.Errorf("Expected country name 'Test Country', got '%s'", details.Name)
    }
}

func TestVehicleUsage(t *testing.T) {
    details, err := GetCountryDetails("Test Country")
    if err != nil {
        t.Fatalf("Failed to get country details: %v", err)
    }

    // Check for test vehicles
    foundVehicle1 := false
    foundVehicle2 := false
    for _, v := range details.Vehicles {
        if v.Vehicle.ID == "1000" {
            foundVehicle1 = true
            if v.Name != "Test Vehicle 1" {
                t.Errorf("Expected vehicle name 'Test Vehicle 1', got '%s'", v.Name)
            }
        } else if v.Vehicle.ID == "1001" {
            foundVehicle2 = true
            if v.Name != "Test Vehicle 2" {
                t.Errorf("Expected vehicle name 'Test Vehicle 2', got '%s'", v.Name)
            }
        }
    }

    if !foundVehicle1 {
        t.Error("Test Vehicle 1 not found")
    }
    if !foundVehicle2 {
        t.Error("Test Vehicle 2 not found")
    }
}

func TestWeaponUsage(t *testing.T) {
    details, err := GetCountryDetails("Test Country")
    if err != nil {
        t.Fatalf("Failed to get country details: %v", err)
    }

    // Check for test weapons
    foundWeapon1 := false
    foundWeapon2 := false
    for _, w := range details.Weapons {
        if w.Weapon.ID == "1000" {
            foundWeapon1 = true
            if w.Name != "Test Rifle" {
                t.Errorf("Expected weapon name 'Test Rifle', got '%s'", w.Name)
            }
        } else if w.Weapon.ID == "1001" {
            foundWeapon2 = true
            if w.Name != "Test Machine Gun" {
                t.Errorf("Expected weapon name 'Test Machine Gun', got '%s'", w.Name)
            }
        }
    }

    if !foundWeapon1 {
        t.Error("Test Rifle not found")
    }
    if !foundWeapon2 {
        t.Error("Test Machine Gun not found")
    }
}