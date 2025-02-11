package tests

import (
	"testing"
	"time"
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
		// First create a weapon and associate it with a member
		_, err := testDB.Exec(`
			INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
			VALUES (101, 'Delete Test Weapon', 'Test Type', '5.56mm')
		`)
		if err != nil {
			t.Fatal("Failed to create test weapon for deletion:", err)
		}

		// Create a test member
		_, err = testDB.Exec(`
			INSERT INTO members (member_id, member_role, member_rank)
			VALUES (101, 'Test Role', 'Test Rank')
		`)
		if err != nil {
			t.Fatal("Failed to create test member:", err)
		}

		// Associate weapon with member
		_, err = testDB.Exec(`
			INSERT INTO members_weapons (member_id, weapon_id)
			VALUES (101, 101)
		`)
		if err != nil {
			t.Fatal("Failed to associate weapon with member:", err)
		}

		// Delete the weapon
		_, err = testDB.Exec(`DELETE FROM weapons WHERE weapon_id = 101`)
		if err != nil {
			t.Fatal("Failed to delete weapon:", err)
		}

		// Verify weapon was deleted
		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM weapons WHERE weapon_id = 101").Scan(&count)
		if err != nil || count != 0 {
			t.Fatal("Weapon deletion verification failed")
		}

		// Verify association was removed
		err = testDB.QueryRow("SELECT COUNT(*) FROM members_weapons WHERE weapon_id = 101").Scan(&count)
		if err != nil || count != 0 {
			t.Fatal("Weapon association cleanup failed")
		}
	})
}

func TestWeaponImageOperations(t *testing.T) {
	if err := resetTestDB(); err != nil {
		t.Fatal("Failed to reset test database:", err)
	}

	t.Run("Upload Weapon Image", func(t *testing.T) {
		// Create test weapon
		_, err := testDB.Exec(`
			INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
			VALUES (102, 'Image Test Weapon', 'Test Type', '5.56mm')
		`)
		if err != nil {
			t.Fatal("Failed to create test weapon:", err)
		}

		testImageData := []byte("fake image data")
		filename := "test-weapon-102.jpg"

		// Store image in mock storage
		if err := mockStorage.Store(filename, testImageData); err != nil {
			t.Fatal("Failed to store test image:", err)
		}

		// Verify image was stored
		storedData, err := mockStorage.Retrieve(filename)
		if err != nil {
			t.Fatal("Failed to retrieve test image:", err)
		}

		if string(storedData) != string(testImageData) {
			t.Fatal("Stored image data does not match original")
		}

		// Update weapon record with mock image URL
		_, err = testDB.Exec(`
			UPDATE weapons 
			SET image_url = ? 
			WHERE weapon_id = 102`,
			"mock://"+filename)
		if err != nil {
			t.Fatal("Failed to update weapon with image URL:", err)
		}

		// Verify image URL was updated
		var imageURL string
		err = testDB.QueryRow("SELECT image_url FROM weapons WHERE weapon_id = 102").Scan(&imageURL)
		if err != nil {
			t.Fatal("Failed to read image URL:", err)
		}

		if imageURL != "mock://"+filename {
			t.Fatal("Image URL mismatch")
		}
	})

	t.Run("Update Weapon Image", func(t *testing.T) {
		testImageData := []byte("updated image data")
		filename := "test-weapon-102-updated.jpg"

		// Store updated image
		if err := mockStorage.Store(filename, testImageData); err != nil {
			t.Fatal("Failed to store updated test image:", err)
		}

		// Update weapon record with new image URL
		_, err := testDB.Exec(`
			UPDATE weapons 
			SET image_url = ? 
			WHERE weapon_id = 102`,
			"mock://"+filename)
		if err != nil {
			t.Fatal("Failed to update weapon with new image URL:", err)
		}

		// Verify new image URL was updated
		var imageURL string
		err = testDB.QueryRow("SELECT image_url FROM weapons WHERE weapon_id = 102").Scan(&imageURL)
		if err != nil {
			t.Fatal("Failed to read updated image URL:", err)
		}

		if imageURL != "mock://"+filename {
			t.Fatal("Updated image URL mismatch")
		}
	})

	t.Run("Delete Weapon with Image", func(t *testing.T) {
		// Get the image URL before deletion
		var imageURL string
		err := testDB.QueryRow("SELECT image_url FROM weapons WHERE weapon_id = 102").Scan(&imageURL)
		if err != nil {
			t.Fatal("Failed to read image URL:", err)
		}

		// Delete the weapon
		_, err = testDB.Exec(`DELETE FROM weapons WHERE weapon_id = 102`)
		if err != nil {
			t.Fatal("Failed to delete weapon:", err)
		}

		// Verify weapon was deleted
		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM weapons WHERE weapon_id = 102").Scan(&count)
		if err != nil || count != 0 {
			t.Fatal("Weapon deletion verification failed")
		}

		// Verify image was deleted from storage
		filename := imageURL[7:] // Remove "mock://" prefix
		_, err = mockStorage.Retrieve(filename)
		if err == nil {
			t.Fatal("Image should have been deleted from storage")
		}
	})
}

func TestWeaponValidation(t *testing.T) {
	if err := resetTestDB(); err != nil {
		t.Fatal("Failed to reset test database:", err)
	}

	t.Run("Prevent Duplicate Weapon IDs", func(t *testing.T) {
		// Create initial weapon
		_, err := testDB.Exec(`
			INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
			VALUES (103, 'Original Weapon', 'Test Type', '5.56mm')
		`)
		if err != nil {
			t.Fatal("Failed to create initial weapon:", err)
		}

		// Attempt to create weapon with duplicate ID
		_, err = testDB.Exec(`
			INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
			VALUES (103, 'Duplicate Weapon', 'Test Type', '5.56mm')
		`)
		if err == nil {
			t.Fatal("Should not allow duplicate weapon IDs")
		}
	})

	t.Run("Required Fields Validation", func(t *testing.T) {
		// Attempt to create weapon with missing required fields
		_, err := testDB.Exec(`
			INSERT INTO weapons (weapon_id)
			VALUES (104)
		`)
		if err == nil {
			t.Fatal("Should not allow weapon creation with missing required fields")
		}
	})

	t.Run("Weapon References Integrity", func(t *testing.T) {
		// Create test weapon and member
		_, err := testDB.Exec(`
			INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
			VALUES (105, 'Reference Test Weapon', 'Test Type', '5.56mm')
		`)
		if err != nil {
			t.Fatal("Failed to create test weapon:", err)
		}

		_, err = testDB.Exec(`
			INSERT INTO members (member_id, member_role, member_rank)
			VALUES (105, 'Test Role', 'Test Rank')
		`)
		if err != nil {
			t.Fatal("Failed to create test member:", err)
		}

		// Create valid reference
		_, err = testDB.Exec(`
			INSERT INTO members_weapons (member_id, weapon_id)
			VALUES (105, 105)
		`)
		if err != nil {
			t.Fatal("Failed to create valid weapon reference:", err)
		}

		// Attempt to reference non-existent weapon
		_, err = testDB.Exec(`
			INSERT INTO members_weapons (member_id, weapon_id)
			VALUES (105, 999)
		`)
		if err == nil {
			t.Fatal("Should not allow references to non-existent weapons")
		}
	})
}

func TestWeaponQueries(t *testing.T) {
	if err := resetTestDB(); err != nil {
		t.Fatal("Failed to reset test database:", err)
	}

	// Set up test data
	setupSQL := `
		INSERT INTO weapons (weapon_id, weapon_name, weapon_type, weapon_caliber)
		VALUES 
			(106, 'Query Test Rifle 1', 'Assault Rifle', '5.56mm'),
			(107, 'Query Test Rifle 2', 'Assault Rifle', '5.56mm'),
			(108, 'Query Test MG', 'Machine Gun', '7.62mm');
	`
	if _, err := testDB.Exec(setupSQL); err != nil {
		t.Fatal("Failed to set up test data:", err)
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
