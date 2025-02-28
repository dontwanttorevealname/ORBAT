package database

import (
	"fmt"
	"net/url"

	"orbat/internal/models"
)

// GetCountries retrieves all countries from the database
func GetCountries() ([]string, error) {
	rows, err := DB.Query(`
		SELECT DISTINCT group_nationality 
		FROM groups 
		ORDER BY group_nationality`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []string
	for rows.Next() {
		var country string
		if err := rows.Scan(&country); err != nil {
			return nil, err
		}
		countries = append(countries, country)
	}
	return countries, nil
}

// GetCountryDetails retrieves detailed information about a country
func GetCountryDetails(countryName string) (models.CountryDetails, error) {
	// URL decode the country name to handle spaces
	decodedName, err := url.QueryUnescape(countryName)
	if err != nil {
		return models.CountryDetails{}, fmt.Errorf("invalid country name: %v", err)
	}

	var details models.CountryDetails
	details.Name = decodedName

	// Update queries to use decoded name
	groups, err := DB.Query(`
		SELECT group_id, group_name, group_nationality, group_size 
		FROM groups 
		WHERE group_nationality = ?
		ORDER BY group_name`, decodedName)
	if err != nil {
		return details, err
	}
	defer groups.Close()

	for groups.Next() {
		var g models.Group
		if err := groups.Scan(&g.ID, &g.Name, &g.Nationality, &g.Size); err != nil {
			return details, err
		}
		details.Groups = append(details.Groups, g)
	}

	// Get weapons used by this country's groups
	weapons, err := DB.Query(`
		SELECT 
			w.weapon_id,
			w.weapon_name,
			w.weapon_type,
			w.weapon_caliber,
			w.image_url,
			COUNT(DISTINCT m.member_id) as user_count
		FROM weapons w
		JOIN members_weapons mw ON w.weapon_id = mw.weapon_id
		JOIN members m ON mw.member_id = m.member_id
		JOIN (
			-- Direct group members
			SELECT m.member_id, g.group_nationality
			FROM members m
			JOIN group_members gm ON m.member_id = gm.member_id
			JOIN groups g ON gm.group_id = g.group_id
			WHERE gm.team_id IS NULL
			UNION ALL
			-- Team members
			SELECT m.member_id, g.group_nationality
			FROM members m
			JOIN team_members tm ON m.member_id = tm.member_id
			JOIN group_members gm ON tm.team_id = gm.team_id
			JOIN groups g ON gm.group_id = g.group_id
			UNION ALL
			-- Vehicle crew members
			SELECT m.member_id, g.group_nationality
			FROM members m
			JOIN vehicle_members vm ON m.member_id = vm.member_id
			JOIN group_vehicles gv ON vm.instance_id = gv.instance_id
			JOIN groups g ON gv.group_id = g.group_id
		) membership ON m.member_id = membership.member_id
		WHERE membership.group_nationality = ?
		GROUP BY w.weapon_id
		ORDER BY w.weapon_name`, decodedName)
	if err != nil {
		return details, err
	}
	defer weapons.Close()

	for weapons.Next() {
		var w models.WeaponUsage
		if err := weapons.Scan(&w.ID, &w.Name, &w.Type, &w.Caliber, &w.ImageURL, &w.UserCount); err != nil {
			return details, err
		}
		details.Weapons = append(details.Weapons, w)
	}

	// Get vehicles used by this country's groups
	vehicles, err := DB.Query(`
		SELECT 
			v.vehicle_id,
			v.vehicle_name,
			v.vehicle_type,
			v.vehicle_armament,
			v.image_url,
			COUNT(DISTINCT gv.instance_id) as instance_count
		FROM vehicles v
		JOIN group_vehicles gv ON v.vehicle_id = gv.vehicle_id
		JOIN groups g ON gv.group_id = g.group_id
		WHERE g.group_nationality = ?
		GROUP BY v.vehicle_id
		ORDER BY v.vehicle_name`, decodedName)
	if err != nil {
		return details, err
	}
	defer vehicles.Close()

	for vehicles.Next() {
		var v models.VehicleUsage
		if err := vehicles.Scan(&v.ID, &v.Name, &v.Type, &v.Armament, &v.ImageURL, &v.InstanceCount); err != nil {
			return details, err
		}
		details.Vehicles = append(details.Vehicles, v)
	}

	return details, nil
} 