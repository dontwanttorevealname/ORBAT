package database

import (
	"fmt"
	"net/url"
	"github.com/biter777/countries"
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

	var countryList []string
	for rows.Next() {
		var countryCode string
		if err := rows.Scan(&countryCode); err != nil {
			return nil, err
		}
		// Convert country code to standardized name
		country := countries.ByName(countryCode)
		if country != countries.Unknown {
			countryList = append(countryList, country.Info().Name)
		} else {
			countryList = append(countryList, countryCode) // Fallback to code
		}
	}
	return countryList, nil
}

// GetCountryDetails retrieves detailed information about a country
func GetCountryDetails(countryName string) (models.CountryDetails, error) {
	// URL decode the country name to handle spaces
	decodedName, err := url.QueryUnescape(countryName)
	if err != nil {
		return models.CountryDetails{}, fmt.Errorf("invalid country name: %v", err)
	}

	// Get the standardized country code
	country := countries.ByName(decodedName)
	if country == countries.Unknown {
		return models.CountryDetails{}, fmt.Errorf("invalid country name: %s", decodedName)
	}

	var details models.CountryDetails
	details.Name = country.Info().Name
	countryCode := country.Info().Alpha2

	// Update queries to use country code
	groups, err := DB.Query(`
		SELECT group_id, group_name, group_nationality, group_size 
		FROM groups 
		WHERE group_nationality = ?
		ORDER BY group_name`, countryCode)
	if err != nil {
		return details, err
	}
	defer groups.Close()

	for groups.Next() {
		var g models.Group
		var gCountryCode string
		if err := groups.Scan(&g.ID, &g.Name, &gCountryCode, &g.Size); err != nil {
			return details, err
		}
		// Convert country code to name for display
		gCountry := countries.ByName(gCountryCode)
		if gCountry != countries.Unknown {
			g.Nationality = gCountry.Info().Name
		} else {
			g.Nationality = gCountryCode
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
		ORDER BY w.weapon_name`, countryCode)
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
		ORDER BY v.vehicle_name`, countryCode)
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

// StandardizeCountryCodes updates all existing country names to their standardized Alpha2 codes
func StandardizeCountryCodes() error {
	// First, get all unique nationalities
	rows, err := DB.Query(`
		SELECT DISTINCT group_nationality 
		FROM groups`)
	if err != nil {
		return fmt.Errorf("failed to query nationalities: %v", err)
	}
	defer rows.Close()

	// For each nationality, convert to standard code if needed
	for rows.Next() {
		var nationality string
		if err := rows.Scan(&nationality); err != nil {
			return fmt.Errorf("failed to scan nationality: %v", err)
		}

		// Try to get standardized country code
		country := countries.ByName(nationality)
		if country != countries.Unknown && country.Info().Alpha2 != nationality {
			// Update all groups with this nationality to use the standard code
			_, err = DB.Exec(`
				UPDATE groups 
				SET group_nationality = ? 
				WHERE group_nationality = ?`,
				country.Info().Alpha2, nationality)
			if err != nil {
				return fmt.Errorf("failed to update nationality %s to %s: %v",
					nationality, country.Info().Alpha2, err)
			}
		}
	}

	return nil
} 