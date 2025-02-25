package database

import (
	"database/sql"
	"fmt"

	"orbat/internal/models"
	"orbat/internal/storage"
)

// GetVehicles retrieves all vehicles from the database
func GetVehicles() ([]models.Vehicle, error) {
	rows, err := DB.Query("SELECT vehicle_id, vehicle_name, vehicle_type, vehicle_armament, image_url FROM vehicles ORDER BY vehicle_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []models.Vehicle
	for rows.Next() {
		var v models.Vehicle
		if err := rows.Scan(&v.ID, &v.Name, &v.Type, &v.Armament, &v.ImageURL); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, nil
}

// GetVehicleDetails retrieves detailed information about a vehicle
func GetVehicleDetails(vehicleID string) (models.VehicleDetails, error) {
	var details models.VehicleDetails

	err := DB.QueryRow(`
		SELECT vehicle_id, vehicle_name, vehicle_type, vehicle_armament, image_url 
		FROM vehicles WHERE vehicle_id = ?`, vehicleID).Scan(
		&details.Vehicle.ID, &details.Vehicle.Name, &details.Vehicle.Type, 
		&details.Vehicle.Armament, &details.Vehicle.ImageURL)
	if err != nil {
		return details, err
	}

	rows, err := DB.Query(`
		SELECT 
			g.group_id,
			g.group_name,
			g.group_nationality,
			m.member_role,
			m.member_rank
		FROM group_vehicles gv
		JOIN groups g ON g.group_id = gv.group_id
		JOIN vehicle_members vm ON vm.instance_id = gv.instance_id
		JOIN members m ON m.member_id = vm.member_id
		WHERE gv.vehicle_id = ?
		ORDER BY g.group_id, m.member_role`, vehicleID)
	if err != nil {
		return details, err
	}
	defer rows.Close()

	var currentGroupUsers models.VehicleGroupUsers
	details.Groups = make([]models.VehicleGroupUsers, 0)
	countries := make(map[string]bool)

	for rows.Next() {
		var groupID int
		var groupName, nationality, role, rank string
		
		err := rows.Scan(&groupID, &groupName, &nationality, &role, &rank)
		if err != nil {
			return details, err
		}

		if currentGroupUsers.GroupID != groupID && currentGroupUsers.GroupID != 0 {
			details.Groups = append(details.Groups, currentGroupUsers)
			currentGroupUsers = models.VehicleGroupUsers{}
		}

		if currentGroupUsers.GroupID == 0 {
			currentGroupUsers.GroupID = groupID
			currentGroupUsers.GroupName = groupName
			currentGroupUsers.Nationality = nationality
			currentGroupUsers.Members = make([]models.VehicleMember, 0)
		}

		currentGroupUsers.Members = append(currentGroupUsers.Members, models.VehicleMember{
			Role: role,
			Rank: rank,
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

// DeleteVehicle deletes a vehicle and its associations
func DeleteVehicle(vehicleID string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get the image URL before deleting the vehicle
	var imageURL sql.NullString
	err = tx.QueryRow("SELECT image_url FROM vehicles WHERE vehicle_id = ?", vehicleID).Scan(&imageURL)
	if err != nil {
		return err
	}

	// Get all instance IDs for this vehicle
	rows, err := tx.Query("SELECT instance_id FROM group_vehicles WHERE vehicle_id = ?", vehicleID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Delete members for each instance
	for rows.Next() {
		var instanceID int
		if err := rows.Scan(&instanceID); err != nil {
			return err
		}
		
		// Delete vehicle members
		_, err = tx.Exec("DELETE FROM vehicle_members WHERE instance_id = ?", instanceID)
		if err != nil {
			return err
		}
	}

	// Delete vehicle instances
	_, err = tx.Exec("DELETE FROM group_vehicles WHERE vehicle_id = ?", vehicleID)
	if err != nil {
		return err
	}

	// Delete the vehicle
	_, err = tx.Exec("DELETE FROM vehicles WHERE vehicle_id = ?", vehicleID)
	if err != nil {
		return err
	}

	// Delete image from GCS if it exists
	if imageURL.Valid && imageURL.String != "" {
		if err := storage.DeleteImage(imageURL.String); err != nil {
			// Log the error but continue with the transaction
			fmt.Printf("Warning: Failed to delete image from storage: %v\n", err)
		}
	}

	return tx.Commit()
} 