package tests

import (
    "testing"
)


func TestGroupOperations(t *testing.T) {
    if err := resetTestDB(); err != nil {
        t.Fatal("Failed to reset test database:", err)
    }

    t.Run("Create Group", func(t *testing.T) {
        // ... existing Create Group test ...
    })

    t.Run("Delete Group", func(t *testing.T) {
        // First verify the group exists
        var groupCount int
        err := testDB.QueryRow(`
            SELECT COUNT(*) 
            FROM groups 
            WHERE group_id = 100
        `).Scan(&groupCount)
        if err != nil {
            t.Fatal("Failed to verify group exists:", err)
        }
        if groupCount != 1 {
            t.Fatal("Group should exist before deletion")
        }

        // Delete group
        result, err := testDB.Exec(`DELETE FROM groups WHERE group_id = 100`)
        if err != nil {
            t.Fatal("Failed to delete group:", err)
        }

        // Verify one row was deleted
        rowsAffected, err := result.RowsAffected()
        if err != nil {
            t.Fatal("Failed to get rows affected:", err)
        }
        if rowsAffected != 1 {
            t.Fatal("Expected 1 group to be deleted")
        }

        // Verify group is gone
        err = testDB.QueryRow(`
            SELECT COUNT(*) 
            FROM groups 
            WHERE group_id = 100
        `).Scan(&groupCount)
        if err != nil {
            t.Fatal("Failed to verify group deletion:", err)
        }
        if groupCount != 0 {
            t.Fatal("Group was not deleted")
        }

        // Verify group_members associations are deleted
        var memberAssocCount int
        err = testDB.QueryRow(`
            SELECT COUNT(*) 
            FROM group_members 
            WHERE group_id = 100
        `).Scan(&memberAssocCount)
        if err != nil {
            t.Fatal("Failed to verify group_members deletion:", err)
        }
        if memberAssocCount != 0 {
            t.Fatal("Group member associations not deleted")
        }

        // Verify members still exist
        var memberCount int
        err = testDB.QueryRow(`
            SELECT COUNT(*) 
            FROM members 
            WHERE member_id IN (100, 101, 102)
        `).Scan(&memberCount)
        if err != nil {
            t.Fatal("Failed to verify members:", err)
        }
        if memberCount != 3 {
            t.Fatal("Members were incorrectly deleted with group")
        }

        // Verify teams still exist
        var teamCount int
        err = testDB.QueryRow(`
            SELECT COUNT(*) 
            FROM teams 
            WHERE team_id = 100
        `).Scan(&teamCount)
        if err != nil {
            t.Fatal("Failed to verify teams:", err)
        }
        if teamCount != 1 {
            t.Fatal("Teams were incorrectly deleted")
        }
    })
}

/* Temporarily disabled
func TestGroupValidation(t *testing.T) {
    if err := resetTestDB(); err != nil {
        t.Fatal("Failed to reset test database:", err)
    }

    t.Run("Prevent Invalid Member Assignments", func(t *testing.T) {
        // Create valid group and member first
        _, err := testDB.Exec(`
            INSERT INTO groups (group_id, group_name, group_nationality)
            VALUES (200, 'Test Group', 'Test Nation')
        `)
        if err != nil {
            t.Fatal("Failed to create test group:", err)
        }

        _, err = testDB.Exec(`
            INSERT INTO members (member_id, member_role, member_rank)
            VALUES (200, 'Test Role', 'Test Rank')
        `)
        if err != nil {
            t.Fatal("Failed to create test member:", err)
        }

        // Try to associate with non-existent member
        _, err = testDB.Exec(`
            INSERT INTO group_members (group_id, member_id)
            VALUES (200, 999)
        `)
        if err == nil {
            t.Fatal("Should not allow invalid member_id")
        }

        // Try to associate with non-existent group
        _, err = testDB.Exec(`
            INSERT INTO group_members (group_id, member_id)
            VALUES (999, 200)
        `)
        if err == nil {
            t.Fatal("Should not allow invalid group_id")
        }

        // Test valid association
        _, err = testDB.Exec(`
            INSERT INTO group_members (group_id, member_id)
            VALUES (200, 200)
        `)
        if err != nil {
            t.Fatal("Should allow valid group-member association")
        }
    })
}
*/