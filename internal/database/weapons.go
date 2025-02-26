package database

import (
	"database/sql"
	"fmt"

	"orbat/internal/models"
	"orbat/internal/storage"
)

// GetWeapons retrieves all weapons from the database
func GetWeapons() ([]models.Weapon, error) {
	rows, err := DB.Query("SELECT weapon_id, weapon_name, weapon_type, weapon_caliber, image_url FROM weapons ORDER BY weapon_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var weapons []models.Weapon
	for rows.Next() {
		var w models.Weapon
		if err := rows.Scan(&w.ID, &w.Name, &w.Type, &w.Caliber, &w.ImageURL); err != nil {
			return nil, err
		}
		weapons = append(weapons, w)
	}
	return weapons, nil
}

// WeaponExists checks if a weapon with the given name exists
func WeaponExists(name string) (bool, int, error) {
	var id int
	err := DB.QueryRow("SELECT weapon_id FROM weapons WHERE weapon_name = ?", name).Scan(&id)
	if err == sql.ErrNoRows {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return true, id, nil
}

// GetWeaponDetails retrieves detailed information about a weapon
func GetWeaponDetails(weaponID string) (models.WeaponDetails, error) {
	var details models.WeaponDetails

	// Get weapon details
	err := DB.QueryRow(`
		SELECT weapon_id, weapon_name, weapon_type, weapon_caliber, image_url 
		FROM weapons WHERE weapon_id = ?`, weaponID).Scan(
		&details.Weapon.ID, &details.Weapon.Name, &details.Weapon.Type, 
		&details.Weapon.Caliber, &details.Weapon.ImageURL)
	if err != nil {
		return details, err
	}

	// Get all users of this weapon and their group info
	rows, err := DB.Query(`
		SELECT 
			g.group_id,
			g.group_name,
			g.group_nationality,
			m.member_role,
			m.member_rank,
			COALESCE(t.team_name, '') as team_name
		FROM members_weapons mw
		JOIN members m ON mw.member_id = m.member_id
		JOIN (
			-- Direct group members
			SELECT member_id, group_id, NULL as team_id 
			FROM group_members 
			WHERE team_id IS NULL
			UNION ALL
			-- Team members
			SELECT tm.member_id, gm.group_id, tm.team_id
			FROM team_members tm
			JOIN group_members gm ON tm.team_id = gm.team_id
			UNION ALL
			-- Vehicle crew members
			SELECT vm.member_id, gv.group_id, NULL as team_id
			FROM vehicle_members vm
			JOIN group_vehicles gv ON vm.instance_id = gv.instance_id
		) membership ON m.member_id = membership.member_id
		JOIN groups g ON membership.group_id = g.group_id
		LEFT JOIN teams t ON membership.team_id = t.team_id
		WHERE mw.weapon_id = ?
		ORDER BY g.group_name, t.team_name`, weaponID)
	if err != nil {
		return details, err
	}
	defer rows.Close()

	var currentGroupUsers models.WeaponGroupUsers
	details.Groups = make([]models.WeaponGroupUsers, 0)
	countries := make(map[string]bool)

	for rows.Next() {
		var groupID int
		var groupName, nationality, role, rank string
		var teamName sql.NullString
		
		err := rows.Scan(&groupID, &groupName, &nationality, &role, &rank, &teamName)
		if err != nil {
			return details, err
		}

		if currentGroupUsers.GroupID != groupID && currentGroupUsers.GroupID != 0 {
			details.Groups = append(details.Groups, currentGroupUsers)
			currentGroupUsers = models.WeaponGroupUsers{}
		}

		if currentGroupUsers.GroupID == 0 {
			currentGroupUsers.GroupID = groupID
			currentGroupUsers.GroupName = groupName
			currentGroupUsers.Nationality = nationality
			currentGroupUsers.Users = make([]models.WeaponUser, 0)
		}

		currentGroupUsers.Users = append(currentGroupUsers.Users, models.WeaponUser{
			Role:     role,
			Rank:     rank,
			TeamName: teamName.String,
		})

		countries[nationality] = true
		details.TotalUsers++
	}

	if currentGroupUsers.GroupID != 0 {
		details.Groups = append(details.Groups, currentGroupUsers)
	}

	details.CountryCount = len(countries)
	details.Countries = make([]string, 0, len(countries))
	for country := range countries {
		details.Countries = append(details.Countries, country)
	}

	return details, nil
}

// DeleteWeapon deletes a weapon and its associations
func DeleteWeapon(weaponID string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the image URL before deleting the weapon
	var imageURL sql.NullString
	err = tx.QueryRow("SELECT image_url FROM weapons WHERE weapon_id = ?", weaponID).Scan(&imageURL)
	if err != nil {
		return err
	}

	// Delete weapon associations first
	_, err = tx.Exec("DELETE FROM members_weapons WHERE weapon_id = ?", weaponID)
	if err != nil {
		return err
	}

	// Delete the weapon itself
	_, err = tx.Exec("DELETE FROM weapons WHERE weapon_id = ?", weaponID)
	if err != nil {
		return err
	}

	// If there was an image, delete it from GCS
	if imageURL.Valid && imageURL.String != "" {
		if err := storage.DeleteImage(imageURL.String); err != nil {
			// Log the error but continue with the transaction
			fmt.Printf("Warning: Failed to delete image from storage: %v\n", err)
		}
	}

	return tx.Commit()
}

// GetMemberWeaponsData retrieves weapons data for a specific member
func GetMemberWeaponsData(memberID string) (map[string]interface{}, error) {
	// Get all available weapons
	allWeapons, err := GetWeapons()
	if err != nil {
		return nil, err
	}

	// Get member's current weapons
	rows, err := DB.Query(`
		SELECT w.weapon_id, w.weapon_name, w.weapon_type, w.weapon_caliber
		FROM members_weapons mw
		JOIN weapons w ON mw.weapon_id = w.weapon_id
		WHERE mw.member_id = ?`, memberID)
	if err != nil {
		// If there's an error querying current weapons, still return all weapons
		// but with an empty current weapons array
		return map[string]interface{}{
			"all":     allWeapons,
			"current": []models.Weapon{},
		}, nil
	}
	defer rows.Close()

	var currentWeapons []models.Weapon
	for rows.Next() {
		var w models.Weapon
		if err := rows.Scan(&w.ID, &w.Name, &w.Type, &w.Caliber); err != nil {
			// If there's an error scanning a weapon, skip it and continue
			continue
		}
		currentWeapons = append(currentWeapons, w)
	}

	// Even if there are no current weapons, we still return a valid response
	return map[string]interface{}{
		"all":     allWeapons,
		"current": currentWeapons,
	}, nil
}

// UpdateMemberWeapons updates the weapons associated with a member
func UpdateMemberWeapons(memberID string, weaponIDs []string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove all existing weapons for this member
	_, err = tx.Exec("DELETE FROM members_weapons WHERE member_id = ?", memberID)
	if err != nil {
		return err
	}

	// If no weapons were selected, just return after deleting existing weapons
	if len(weaponIDs) == 0 {
		return tx.Commit()
	}

	// Add new weapons
	for _, weaponID := range weaponIDs {
		// Verify the weapon exists before inserting
		var exists bool
		err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM weapons WHERE weapon_id = ?)", weaponID).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			continue // Skip weapons that don't exist
		}

		_, err = tx.Exec(
			"INSERT INTO members_weapons (member_id, weapon_id) VALUES (?, ?)",
			memberID, weaponID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
} 