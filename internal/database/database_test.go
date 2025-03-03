package database

import (
    "os"
    "testing"
    "github.com/joho/godotenv"
    "fmt"
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

    // Debug: Print all countries
    t.Logf("Found countries: %v", countries)

    // Check for test country
    testCountryFound := false
    for _, country := range countries {
        if country == "Test Nation" { // Changed: Match the actual value from seed file
            testCountryFound = true
            break
        }
    }
    if !testCountryFound {
        t.Error("Test Nation not found in countries list")
    }

    // Test country details
    details, err := GetCountryDetails("Test Nation") // Changed: Match the actual value
    if err != nil {
        t.Fatalf("Failed to get country details: %v", err)
    }

    if details.Name != "Test Nation" { // Changed: Match the actual value
        t.Errorf("Expected country name 'Test Nation', got '%s'", details.Name)
    }
}

func TestVehicleUsage(t *testing.T) {
    details, err := GetCountryDetails("Test Nation") // Changed: Match the actual value
    if err != nil {
        t.Fatalf("Failed to get country details: %v", err)
    }

    // Debug: Print all vehicles
    t.Logf("Found vehicles: %+v", details.Vehicles)

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
    details, err := GetCountryDetails("Test Nation") // Changed: Match the actual value
    if err != nil {
        t.Fatalf("Failed to get country details: %v", err)
    }

    // Debug: Print all weapons
    t.Logf("Found weapons: %+v", details.Weapons)

    // Check for test weapons
    foundWeapon1 := false
    foundWeapon2 := false
    for _, w := range details.Weapons {
        if w.Weapon.ID == 1000 {
            foundWeapon1 = true
            if w.Name != "Test Rifle" {
                t.Errorf("Expected weapon name 'Test Rifle', got '%s'", w.Name)
            }
        } else if w.Weapon.ID == 1001 {
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

func TestCreateAndDeleteWeapon(t *testing.T) {
    // Create a new test weapon
    newWeapon := models.Weapon{
        ID:      2000,
        Name:    "Test Create Weapon",
        Type:    "Test Create Type",
        Caliber: "Test Create Caliber",
    }

    // Insert the weapon directly using SQL
    _, err := DB.Exec(`
        INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
        VALUES (?, ?, ?, ?)`,
        newWeapon.ID, newWeapon.Name, newWeapon.Type, newWeapon.Caliber)
    if err != nil {
        t.Fatalf("Failed to create weapon: %v", err)
    }

    // Verify the weapon was created
    weapons, err := GetWeapons()
    if err != nil {
        t.Fatalf("Failed to get weapons: %v", err)
    }

    found := false
    for _, w := range weapons {
        if w.ID == newWeapon.ID {
            found = true
            if w.Name != newWeapon.Name {
                t.Errorf("Expected weapon name '%s', got '%s'", newWeapon.Name, w.Name)
            }
            break
        }
    }
    if !found {
        t.Error("Created weapon not found in database")
    }

    // Cleanup
    err = DeleteWeapon(fmt.Sprintf("%d", newWeapon.ID))
    if err != nil {
        t.Fatalf("Failed to cleanup test weapon: %v", err)
    }
}

func TestCreateAndDeleteGroup(t *testing.T) {
    // Create a new test group
    groupName := "Test Create Group"
    nationality := "Test Nation"
    
    // Insert the group directly using SQL
    result, err := DB.Exec(`
        INSERT INTO groups (group_name, group_nationality, group_size)
        VALUES (?, ?, 0)`,
        groupName, nationality)
    if err != nil {
        t.Fatalf("Failed to create group: %v", err)
    }

    groupID, err := result.LastInsertId()
    if err != nil {
        t.Fatalf("Failed to get last insert ID: %v", err)
    }

    // Verify the group was created
    groups, err := GetGroups()
    if err != nil {
        t.Fatalf("Failed to get groups: %v", err)
    }

    found := false
    for _, g := range groups {
        if g.ID == int(groupID) {
            found = true
            if g.Name != groupName {
                t.Errorf("Expected group name '%s', got '%s'", groupName, g.Name)
            }
            break
        }
    }
    if !found {
        t.Error("Created group not found in database")
    }

    // Cleanup
    err = DeleteGroup(DB, fmt.Sprintf("%d", groupID))
    if err != nil {
        t.Fatalf("Failed to cleanup test group: %v", err)
    }
}

func TestUpdateWeapon(t *testing.T) {
    // Create a test weapon
    initialWeapon := models.Weapon{
        ID:      3000,
        Name:    "Initial Weapon",
        Type:    "Initial Type",
        Caliber: "Initial Caliber",
    }

    // Insert the weapon directly using SQL
    _, err := DB.Exec(`
        INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
        VALUES (?, ?, ?, ?)`,
        initialWeapon.ID, initialWeapon.Name, initialWeapon.Type, initialWeapon.Caliber)
    if err != nil {
        t.Fatalf("Failed to create initial weapon: %v", err)
    }

    // Update the weapon directly using SQL
    updatedName := "Updated Weapon"
    updatedType := "Updated Type"
    _, err = DB.Exec(`
        UPDATE weapons 
        SET weapon_name = ?, weapon_type = ?
        WHERE weapon_id = ?`,
        updatedName, updatedType, initialWeapon.ID)
    if err != nil {
        t.Fatalf("Failed to update weapon: %v", err)
    }

    // Verify the update
    weapons, err := GetWeapons()
    if err != nil {
        t.Fatalf("Failed to get weapons: %v", err)
    }

    found := false
    for _, w := range weapons {
        if w.ID == initialWeapon.ID {
            found = true
            if w.Name != updatedName {
                t.Errorf("Expected updated weapon name '%s', got '%s'", updatedName, w.Name)
            }
            if w.Type != updatedType {
                t.Errorf("Expected updated weapon type '%s', got '%s'", updatedType, w.Type)
            }
            break
        }
    }
    if !found {
        t.Error("Updated weapon not found in database")
    }

    // Cleanup
    err = DeleteWeapon(fmt.Sprintf("%d", initialWeapon.ID))
    if err != nil {
        t.Fatalf("Failed to cleanup test weapon: %v", err)
    }
}