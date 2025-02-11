package tests

import (
	"testing"
)

func TestGroupOperations(t *testing.T) {
	if err := resetTestDB(); err != nil {
		t.Fatal("Failed to reset test database:", err)
	}

	t.Run("Create Group with Members", func(t *testing.T) {
		tx, err := testDB.Begin()
		if err != nil {
			t.Fatal("Failed to begin transaction:", err)
		}
		defer tx.Rollback()

		// Create a new group
		_, err = tx.Exec(`
			INSERT INTO groups (group_id, group_name, group_nationality)
			VALUES (100, 'Test Squad', 'Test Nation')
		`)
		if err != nil {
			t.Fatal("Failed to create group:", err)
		}

		// Create members
		_, err = tx.Exec(`
			INSERT INTO members (member_id, member_role, member_rank)
			VALUES 
				(100, 'Squad Leader', 'Sergeant'),
				(101, 'Rifleman', 'Private')
		`)
		if err != nil {
			t.Fatal("Failed to create members:", err)
		}

		// Create team
		_, err = tx.Exec(`
			INSERT INTO teams (team_id, team_name)
			VALUES (100, 'Test Team')
		`)
		if err != nil {
			t.Fatal("Failed to create team:", err)
		}

		// Associate members with group
		_, err = tx.Exec(`
			INSERT INTO group_members (group_id, member_id, team_id)
			VALUES 
				(100, 100, NULL),
				(100, NULL, 100)
		`)
		if err != nil {
			t.Fatal("Failed to associate members with group:", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatal("Failed to commit transaction:", err)
		}

		// Verify group structure
		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM group_members WHERE group_id = 100").Scan(&count)
		if err != nil || count != 2 {
			t.Fatal("Group structure verification failed")
		}
	})
}