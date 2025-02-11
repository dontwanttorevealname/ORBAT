package tests

import (
    "testing"
)

func TestGroupOperations(t *testing.T) {
    if err := resetTestDB(); err != nil {
        t.Fatal("Failed to reset test database:", err)
    }

    t.Run("Create Group", func(t *testing.T) {
        tx, err := testDB.Begin()
        if err != nil {
            t.Fatal("Failed to begin transaction:", err)
        }

        // Create group
        _, err = tx.Exec(`
            INSERT INTO groups (group_id, group_name, group_nationality)
            VALUES (100, 'Test Squad', 'Test Nation')
        `)
        if err != nil {
            tx.Rollback()
            t.Fatal("Failed to create group:", err)
        }

        // Create members
        _, err = tx.Exec(`
            INSERT INTO members (member_id, member_role, member_rank)
            VALUES 
                (100, 'Squad Leader', 'Sergeant'),
                (101, 'Team Leader', 'Corporal'),
                (102, 'Rifleman', 'Private')
        `)
        if err != nil {
            tx.Rollback()
            t.Fatal("Failed to create members:", err)
        }

        // Create team
        _, err = tx.Exec(`
            INSERT INTO teams (team_id, team_name)
            VALUES (100, 'Alpha Team')
        `)
        if err != nil {
            tx.Rollback()
            t.Fatal("Failed to create team:", err)
        }

        // Associate members with group and team
        _, err = tx.Exec(`
            INSERT INTO group_members (group_id, member_id, team_id)
            VALUES 
                (100, 100, NULL),  -- Squad Leader (no team)
                (100, 101, 100),   -- Team Leader in Alpha Team
                (100, 102, 100)    -- Rifleman in Alpha Team
        `)
        if err != nil {
            tx.Rollback()
            t.Fatal("Failed to associate members:", err)
        }

        if err := tx.Commit(); err != nil {
            t.Fatal("Failed to commit transaction:", err)
        }

        // Verify group structure
        var memberCount int
        err = testDB.QueryRow(`
            SELECT COUNT(*) 
            FROM group_members 
            WHERE group_id = 100
        `).Scan(&memberCount)
        
        if err != nil {
            t.Fatal("Failed to verify group members:", err)
        }

        if memberCount != 3 {
            t.Fatalf("Expected 3 group members, got %d", memberCount)
        }
    })

    t.Run("Delete Group", func(t *testing.T) {
        // Delete group should cascade to group_members
        _, err := testDB.Exec(`DELETE FROM groups WHERE group_id = 100`)
        if err != nil {
            t.Fatal("Failed to delete group:", err)
        }

        // Verify group and associations are deleted
        var count int
        err = testDB.QueryRow(`
            SELECT COUNT(*) 
            FROM group_members 
            WHERE group_id = 100
        `).Scan(&count)
        
        if err != nil {
            t.Fatal("Failed to verify group deletion:", err)
        }

        if count != 0 {
            t.Fatal("Group member associations not deleted")
        }

        // Members and teams should still exist
        err = testDB.QueryRow(`
            SELECT COUNT(*) 
            FROM members 
            WHERE member_id IN (100, 101, 102)
        `).Scan(&count)
        
        if err != nil {
            t.Fatal("Failed to verify members:", err)
        }

        if count != 3 {
            t.Fatal("Members should not be deleted with group")
        }
    })
}

func TestGroupValidation(t *testing.T) {
    if err := resetTestDB(); err != nil {
        t.Fatal("Failed to reset test database:", err)
    }

    t.Run("Prevent Invalid Member Assignments", func(t *testing.T) {
        // Try to create group_members entry with non-existent group
        _, err := testDB.Exec(`
            INSERT INTO group_members (group_id, member_id)
            VALUES (999, 1)
        `)
        if err == nil {
            t.Fatal("Should not allow invalid group_id")
        }

        // Try to create group_members entry with non-existent member
        _, err = testDB.Exec(`
            INSERT INTO groups (group_id, group_name, group_nationality)
            VALUES (200, 'Test Group', 'Test Nation')
        `)
        if err != nil {
            t.Fatal("Failed to create test group:", err)
        }

        _, err = testDB.Exec(`
            INSERT INTO group_members (group_id, member_id)
            VALUES (200, 999)
        `)
        if err == nil {
            t.Fatal("Should not allow invalid member_id")
        }
    })

    t.Run("Required Fields", func(t *testing.T) {
        _, err := testDB.Exec(`
            INSERT INTO groups (group_id)
            VALUES (201)
        `)
        if err == nil {
            t.Fatal("Should require group_name")
        }
    })
}