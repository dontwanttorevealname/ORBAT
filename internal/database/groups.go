package database

import (
	"database/sql"
	"fmt"

	"orbat/internal/models"
	"github.com/biter777/countries"
)

// GetGroups retrieves all groups from the database
func GetGroups() ([]models.Group, error) {
	rows, err := DB.Query(`
		SELECT 
			g.group_id,
			g.group_name,
			g.group_nationality,
			g.group_size
		FROM groups g
		ORDER BY g.group_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var g models.Group
		var countryCode string
		if err := rows.Scan(&g.ID, &g.Name, &countryCode, &g.Size); err != nil {
			return nil, err
		}
		// Convert country code to name
		country := countries.ByName(countryCode)
		if country != countries.Unknown {
			g.Nationality = country.Info().Name
		} else {
			g.Nationality = countryCode // Fallback to code if conversion fails
		}
		groups = append(groups, g)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

// GetGroupDetails retrieves detailed information about a group
func GetGroupDetails(groupID string) (models.GroupDetails, error) {
	var group models.GroupDetails
	var countryCode string
	
	// Get basic group info
	err := DB.QueryRow(`
		SELECT g.group_id, g.group_name, g.group_size, g.group_nationality 
		FROM groups g 
		WHERE g.group_id = ?`, groupID).Scan(&group.ID, &group.Name, &group.Size, &countryCode)
	if err != nil {
		return group, fmt.Errorf("failed to get group details: %v", err)
	}

	// Convert country code to name
	country := countries.ByName(countryCode)
	if country != countries.Unknown {
		group.Nationality = country.Info().Name
	} else {
		group.Nationality = countryCode // Fallback to code if conversion fails
	}

	// Get direct members (excluding team members and vehicle crew)
	memberRows, err := DB.Query(`
		SELECT DISTINCT m.member_id, m.member_role, m.member_rank
		FROM members m
		JOIN group_members gm ON m.member_id = gm.member_id
		WHERE gm.group_id = ? AND gm.team_id IS NULL`, groupID)
	if err != nil {
		return group, fmt.Errorf("failed to get direct members: %v", err)
	}
	defer memberRows.Close()

	for memberRows.Next() {
		var m models.Member
		err := memberRows.Scan(&m.ID, &m.Role, &m.Rank)
		if err != nil {
			return group, fmt.Errorf("failed to scan member: %v", err)
		}

		// Get member's weapons
		weaponRows, err := DB.Query(`
			SELECT w.weapon_id, w.weapon_name, w.weapon_type, w.weapon_caliber
			FROM weapons w
			JOIN members_weapons mw ON w.weapon_id = mw.weapon_id
			WHERE mw.member_id = ?`, m.ID)
		if err != nil {
			return group, fmt.Errorf("failed to get member weapons: %v", err)
		}
		defer weaponRows.Close()

		for weaponRows.Next() {
			var w models.Weapon
			err := weaponRows.Scan(&w.ID, &w.Name, &w.Type, &w.Caliber)
			if err != nil {
				return group, fmt.Errorf("failed to scan weapon: %v", err)
			}
			m.Weapons = append(m.Weapons, w)
		}

		group.DirectMembers = append(group.DirectMembers, m)
	}

	// Get teams and their members
	teamRows, err := DB.Query(`
		SELECT DISTINCT t.team_id, t.team_name, t.team_size
		FROM teams t
		JOIN group_members gm ON t.team_id = gm.team_id
		WHERE gm.group_id = ?`, groupID)
	if err != nil {
		return group, fmt.Errorf("failed to get teams: %v", err)
	}
	defer teamRows.Close()

	for teamRows.Next() {
		var team models.Team
		err := teamRows.Scan(&team.ID, &team.Name, &team.Size)
		if err != nil {
			return group, fmt.Errorf("failed to scan team: %v", err)
		}

		// Get team members
		teamMemberRows, err := DB.Query(`
			SELECT m.member_id, m.member_role, m.member_rank
			FROM members m
			JOIN team_members tm ON m.member_id = tm.member_id
			WHERE tm.team_id = ?`, team.ID)
		if err != nil {
			return group, fmt.Errorf("failed to get team members: %v", err)
		}
		defer teamMemberRows.Close()

		for teamMemberRows.Next() {
			var m models.Member
			err := teamMemberRows.Scan(&m.ID, &m.Role, &m.Rank)
			if err != nil {
				return group, fmt.Errorf("failed to scan team member: %v", err)
			}

			// Get member's weapons
			weaponRows, err := DB.Query(`
				SELECT w.weapon_id, w.weapon_name, w.weapon_type, w.weapon_caliber
				FROM weapons w
				JOIN members_weapons mw ON w.weapon_id = mw.weapon_id
				WHERE mw.member_id = ?`, m.ID)
			if err != nil {
				return group, fmt.Errorf("failed to get team member weapons: %v", err)
			}
			defer weaponRows.Close()

			for weaponRows.Next() {
				var w models.Weapon
				err := weaponRows.Scan(&w.ID, &w.Name, &w.Type, &w.Caliber)
				if err != nil {
					return group, fmt.Errorf("failed to scan weapon: %v", err)
				}
				m.Weapons = append(m.Weapons, w)
			}

			team.Members = append(team.Members, m)
		}

		group.Teams = append(group.Teams, team)
	}

	// Get vehicles and their crew
	vehicleRows, err := DB.Query(`
		SELECT DISTINCT v.vehicle_id, v.vehicle_name, v.vehicle_type, v.vehicle_armament, v.image_url,
			   gv.instance_id
		FROM vehicles v
		JOIN group_vehicles gv ON v.vehicle_id = gv.vehicle_id
		WHERE gv.group_id = ?`, groupID)
	if err != nil {
		return group, fmt.Errorf("failed to get vehicles: %v", err)
	}
	defer vehicleRows.Close()

	for vehicleRows.Next() {
		var vehicle models.Vehicle
		var instanceID string
		err := vehicleRows.Scan(&vehicle.ID, &vehicle.Name, &vehicle.Type, &vehicle.Armament, &vehicle.ImageURL, &instanceID)
		if err != nil {
			return group, fmt.Errorf("failed to scan vehicle: %v", err)
		}

		// Get vehicle crew members for this specific vehicle instance
		crewRows, err := DB.Query(`
			SELECT DISTINCT m.member_id, m.member_role, m.member_rank
			FROM members m
			JOIN vehicle_members vm ON m.member_id = vm.member_id
			WHERE vm.instance_id = ?`, instanceID)
		if err != nil {
			return group, fmt.Errorf("failed to get vehicle crew: %v", err)
		}
		defer crewRows.Close()

		for crewRows.Next() {
			var m models.Member
			err := crewRows.Scan(&m.ID, &m.Role, &m.Rank)
			if err != nil {
				return group, fmt.Errorf("failed to scan crew member: %v", err)
			}

			// Get crew member's weapons
			weaponRows, err := DB.Query(`
				SELECT w.weapon_id, w.weapon_name, w.weapon_type, w.weapon_caliber
				FROM weapons w
				JOIN members_weapons mw ON w.weapon_id = mw.weapon_id
				WHERE mw.member_id = ?`, m.ID)
			if err != nil {
				return group, fmt.Errorf("failed to get crew weapons: %v", err)
			}
			defer weaponRows.Close()

			for weaponRows.Next() {
				var w models.Weapon
				err := weaponRows.Scan(&w.ID, &w.Name, &w.Type, &w.Caliber)
				if err != nil {
					return group, fmt.Errorf("failed to scan weapon: %v", err)
				}
				m.Weapons = append(m.Weapons, w)
			}

			vehicle.Crew = append(vehicle.Crew, m)
		}

		group.Vehicles = append(group.Vehicles, vehicle)
	}

	return group, nil
}

// DbOrTx is an interface that can be satisfied by either *sql.DB or *sql.Tx
type DbOrTx interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// DeleteGroup deletes a group and all its associated data
func DeleteGroup(db DbOrTx, groupID string) error {
	// 1. Get all member IDs (direct, team, and vehicle members)
	memberIDs := make(map[string]bool)

	// Get direct member IDs
	rows, err := db.Query(`
		SELECT member_id 
		FROM group_members 
		WHERE group_id = ? AND member_id IS NOT NULL`, groupID)
	if err != nil {
		return fmt.Errorf("failed to get member IDs: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var memberID string
		if err := rows.Scan(&memberID); err != nil {
			return fmt.Errorf("failed to scan member ID: %v", err)
		}
		memberIDs[memberID] = true
	}

	// Get team member IDs
	teamRows, err := db.Query(`
		SELECT tm.member_id
		FROM team_members tm
		JOIN group_members gm ON tm.team_id = gm.team_id
		WHERE gm.group_id = ?`, groupID)
	if err != nil {
		return fmt.Errorf("failed to get team member IDs: %v", err)
	}
	defer teamRows.Close()

	for teamRows.Next() {
		var memberID string
		if err := teamRows.Scan(&memberID); err != nil {
			return fmt.Errorf("failed to scan team member ID: %v", err)
		}
		memberIDs[memberID] = true
	}

	// Get vehicle member IDs
	vehicleRows, err := db.Query(`
		SELECT vm.member_id
		FROM vehicle_members vm
		JOIN group_vehicles gv ON vm.instance_id = gv.instance_id
		WHERE gv.group_id = ?`, groupID)
	if err != nil {
		return fmt.Errorf("failed to get vehicle member IDs: %v", err)
	}
	defer vehicleRows.Close()

	for vehicleRows.Next() {
		var memberID string
		if err := vehicleRows.Scan(&memberID); err != nil {
			return fmt.Errorf("failed to scan vehicle member ID: %v", err)
		}
		memberIDs[memberID] = true
	}

	// 2. Delete weapon associations
	for memberID := range memberIDs {
		_, err = db.Exec("DELETE FROM members_weapons WHERE member_id = ?", memberID)
		if err != nil {
			return fmt.Errorf("failed to delete weapon associations: %v", err)
		}
	}

	// 3. Delete vehicle member associations
	_, err = db.Exec(`
		DELETE FROM vehicle_members 
		WHERE instance_id IN (
			SELECT instance_id 
			FROM group_vehicles 
			WHERE group_id = ?
		)`, groupID)
	if err != nil {
		return fmt.Errorf("failed to delete vehicle members: %v", err)
	}

	// 4. Delete group vehicle associations
	_, err = db.Exec("DELETE FROM group_vehicles WHERE group_id = ?", groupID)
	if err != nil {
		return fmt.Errorf("failed to delete group vehicles: %v", err)
	}

	// 5. Delete team member associations
	_, err = db.Exec(`
		DELETE FROM team_members 
		WHERE team_id IN (
			SELECT team_id 
			FROM group_members 
			WHERE group_id = ? AND team_id IS NOT NULL
		)`, groupID)
	if err != nil {
		return fmt.Errorf("failed to delete team members: %v", err)
	}

	// 6. Delete group member associations
	_, err = db.Exec("DELETE FROM group_members WHERE group_id = ?", groupID)
	if err != nil {
		return fmt.Errorf("failed to delete group members: %v", err)
	}

	// 7. Delete members
	for memberID := range memberIDs {
		_, err = db.Exec("DELETE FROM members WHERE member_id = ?", memberID)
		if err != nil {
			return fmt.Errorf("failed to delete members: %v", err)
		}
	}

	// 8. Delete teams
	_, err = db.Exec(`
		DELETE FROM teams 
		WHERE team_id IN (
			SELECT DISTINCT team_id 
			FROM group_members 
			WHERE group_id = ? AND team_id IS NOT NULL
		)`, groupID)
	if err != nil {
		return fmt.Errorf("failed to delete teams: %v", err)
	}

	// 9. Finally delete the group
	_, err = db.Exec("DELETE FROM groups WHERE group_id = ?", groupID)
	if err != nil {
		return fmt.Errorf("failed to delete group: %v", err)
	}

	return nil
} 