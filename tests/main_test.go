package tests

import (
	"testing"
)

var mockStorage = NewMockStorage()

func TestWeaponCRUD(t *testing.T) {
	// Reset database before test
	if err := resetTestDB(); err != nil {
		t.Fatal("Failed to reset test database:", err)
	}

	t.Run("Create Weapon", func(t *testing.T) {
		// Test weapon creation
		_, err := testDB.Exec(`
			INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
			VALUES (100, 'Test Weapon', 'Test Type', '5.56mm')
		`)
		if err != nil {
			t.Fatal("Failed to create weapon:", err)
		}

		// Verify weapon was created
		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM weapons WHERE weapon_id = 100").Scan(&count)
		if err != nil || count != 1 {
			t.Fatal("Weapon creation verification failed")
		}
	})

	t.Run("Read Weapon", func(t *testing.T) {
		var (
			weaponName   string
			weaponType   string
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

		// Verify update
		var (
			weaponName   string
			weaponType   string
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

		// Verify deletion
		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM weapons WHERE weapon_id = 100").Scan(&count)
		if err != nil || count != 0 {
			t.Fatal("Weapon deletion verification failed")
		}
	})

	t.Run("Handle Image Upload", func(t *testing.T) {
		// Create a weapon for image testing
		_, err := testDB.Exec(`
			INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
			VALUES (101, 'Image Test Weapon', 'Test Type', '5.56mm')
		`)
		if err != nil {
			t.Fatal("Failed to create weapon for image test:", err)
		}

		// Test image data
		testImageData := []byte("fake image data")
		filename := "weapon-101.jpg"

		// Store image in mock storage
		if err := mockStorage.Store(filename, testImageData); err != nil {
			t.Fatal("Failed to store test image:", err)
		}

		// Update weapon with image URL
		_, err = testDB.Exec(`
			UPDATE weapons 
			SET image_url = ? 
			WHERE weapon_id = 101`,
			"mock://"+filename)
		if err != nil {
			t.Fatal("Failed to update weapon with image URL:", err)
		}

		// Verify image storage
		storedData, err := mockStorage.Retrieve(filename)
		if err != nil {
			t.Fatal("Failed to retrieve stored image:", err)
		}

		if string(storedData) != string(testImageData) {
			t.Fatal("Stored image data does not match original")
		}
	})

	t.Run("Multiple Weapons Management", func(t *testing.T) {
		// Setup test data
		weapons := []struct {
			id      int
			name    string
			typ     string
			caliber string
		}{
			{106, "Test Rifle 1", "Assault Rifle", "5.56mm"},
			{107, "Test Rifle 2", "Assault Rifle", "5.56mm"},
			{108, "Test MG", "Machine Gun", "7.62mm"},
		}

		// Insert test weapons
		for _, w := range weapons {
			_, err := testDB.Exec(`
				INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
				VALUES (?, ?, ?, ?)
			`, w.id, w.name, w.typ, w.caliber)
			if err != nil {
				t.Fatal("Failed to set up test data:", err)
			}
		}

		t.Run("Filter Weapons by Type", func(t *testing.T) {
			var count int
			err := testDB.QueryRow(`
				SELECT COUNT(*) 
				FROM weapons 
				WHERE weapon_type = 'Assault Rifle' 
				AND weapon_id IN (106, 107, 108)
			`).Scan(&count)
			
			if err != nil {
				t.Fatal("Failed to query weapons by type:", err)
			}

			if count != 2 {
				t.Fatalf("Expected 2 assault rifles, got %d", count)
			}
		})

		t.Run("Filter Weapons by Caliber", func(t *testing.T) {
			var count int
			err := testDB.QueryRow(`
				SELECT COUNT(*) 
				FROM weapons 
				WHERE weapon_caliber = '7.62mm'
				AND weapon_id IN (106, 107, 108)
			`).Scan(&count)
			
			if err != nil {
				t.Fatal("Failed to query weapons by caliber:", err)
			}

			if count != 1 {
				t.Fatalf("Expected 1 7.62mm weapon, got %d", count)
			}
		})
	})
}

func TestWeaponTransactions(t *testing.T) {
	if err := resetTestDB(); err != nil {
		t.Fatal("Failed to reset test database:", err)
	}

	t.Run("Atomic Weapon Creation", func(t *testing.T) {
		tx, err := testDB.Begin()
		if err != nil {
			t.Fatal("Failed to begin transaction:", err)
		}

		// Create weapon within transaction
		_, err = tx.Exec(`
			INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
			VALUES (109, 'Transaction Test Weapon', 'Test Type', '5.56mm')
		`)
		if err != nil {
			tx.Rollback()
			t.Fatal("Failed to create weapon in transaction:", err)
		}

		// Intentionally cause an error to test rollback
		_, err = tx.Exec(`
			INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
			VALUES (109, 'Duplicate Weapon', 'Test Type', '5.56mm')
		`)
		if err == nil {
			tx.Commit()
			t.Fatal("Should not allow duplicate weapon creation")
		}

		// Rollback the transaction
		tx.Rollback()

		// Verify no weapons were created
		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM weapons WHERE weapon_id = 109").Scan(&count)
		if err != nil {
			t.Fatal("Failed to verify weapon count:", err)
		}

		if count != 0 {
			t.Fatal("Transaction rollback failed")
		}
	})
}