package tests

import (
    "testing"
)

func TestWeaponCRUD(t *testing.T) {
    mockStorage := NewMockStorage()
    
    if err := resetTestDB(); err != nil {
        t.Fatal("Failed to reset test database:", err)
    }

    t.Run("Create Weapon", func(t *testing.T) {
        _, err := testDB.Exec(`
            INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
            VALUES (100, 'Test Weapon', 'Test Type', '5.56mm')
        `)
        if err != nil {
            t.Fatal("Failed to create weapon:", err)
        }

        var count int
        err = testDB.QueryRow("SELECT COUNT(*) FROM weapons WHERE weapon_id = 100").Scan(&count)
        if err != nil || count != 1 {
            t.Fatal("Weapon creation verification failed")
        }
    })

    t.Run("Read Weapon", func(t *testing.T) {
        var (
            weaponName    string
            weaponType    string
            weaponCaliber string
        )
        
        err := testDB.QueryRow(`
            SELECT weapon_name, weapon_type, weapon_caliber 
            FROM weapons 
            WHERE weapon_id = 100
        `).Scan(&weaponName, &weaponType, &weaponCaliber)
        
        if err != nil {
            t.Fatal("Failed to read weapon:", err)
        }

        if weaponName != "Test Weapon" || weaponType != "Test Type" || weaponCaliber != "5.56mm" {
            t.Fatalf("Weapon data mismatch. Got: %s, %s, %s", weaponName, weaponType, weaponCaliber)
        }
    })

    t.Run("Update Weapon", func(t *testing.T) {
        _, err := testDB.Exec(`
            UPDATE weapons 
            SET weapon_name = 'Updated Weapon',
                weapon_type = 'Updated Type',
                weapon_caliber = '7.62mm'
            WHERE weapon_id = 100
        `)
        if err != nil {
            t.Fatal("Failed to update weapon:", err)
        }

        var (
            weaponName    string
            weaponType    string
            weaponCaliber string
        )
        
        err = testDB.QueryRow(`
            SELECT weapon_name, weapon_type, weapon_caliber 
            FROM weapons 
            WHERE weapon_id = 100
        `).Scan(&weaponName, &weaponType, &weaponCaliber)
        
        if err != nil {
            t.Fatal("Failed to read updated weapon:", err)
        }

        if weaponName != "Updated Weapon" || weaponType != "Updated Type" || weaponCaliber != "7.62mm" {
            t.Fatal("Weapon update verification failed")
        }
    })

    t.Run("Delete Weapon", func(t *testing.T) {
        _, err := testDB.Exec("DELETE FROM weapons WHERE weapon_id = 100")
        if err != nil {
            t.Fatal("Failed to delete weapon:", err)
        }

        var count int
        err = testDB.QueryRow("SELECT COUNT(*) FROM weapons WHERE weapon_id = 100").Scan(&count)
        if err != nil || count != 0 {
            t.Fatal("Weapon deletion verification failed")
        }
    })

    t.Run("Handle Image Upload", func(t *testing.T) {
        _, err := testDB.Exec(`
            INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
            VALUES (101, 'Image Test Weapon', 'Test Type', '5.56mm')
        `)
        if err != nil {
            t.Fatal("Failed to create weapon for image test:", err)
        }

        testImageData := []byte("fake image data")
        filename := "weapon-101.jpg"

        if err := mockStorage.Store(filename, testImageData); err != nil {
            t.Fatal("Failed to store test image:", err)
        }

        _, err = testDB.Exec(`
            UPDATE weapons 
            SET image_url = ? 
            WHERE weapon_id = 101`,
            "mock://"+filename)
        if err != nil {
            t.Fatal("Failed to update weapon with image URL:", err)
        }

        storedData, err := mockStorage.Retrieve(filename)
        if err != nil {
            t.Fatal("Failed to retrieve stored image:", err)
        }

        if string(storedData) != string(testImageData) {
            t.Fatal("Stored image data does not match original")
        }
    })
}

func TestWeaponValidation(t *testing.T) {
    if err := resetTestDB(); err != nil {
        t.Fatal("Failed to reset test database:", err)
    }

    t.Run("Required Fields", func(t *testing.T) {
        _, err := testDB.Exec(`
            INSERT INTO weapons (weapon_id, weapon_name)
            VALUES (201, 'Incomplete Weapon')
        `)
        if err == nil {
            t.Fatal("Should require weapon_type and weapon_caliber")
        }

        // Try with missing caliber
        _, err = testDB.Exec(`
            INSERT INTO weapons (weapon_id, weapon_name, weapon_type)
            VALUES (201, 'Incomplete Weapon', 'Test Type')
        `)
        if err == nil {
            t.Fatal("Should require weapon_caliber")
        }

        // Try with missing type
        _, err = testDB.Exec(`
            INSERT INTO weapons (weapon_id, weapon_name, weapon_caliber)
            VALUES (201, 'Incomplete Weapon', '5.56mm')
        `)
        if err == nil {
            t.Fatal("Should require weapon_type")
        }
    })
}