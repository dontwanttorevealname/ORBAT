package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"orbat/internal/database"
)

// GroupsHandler handles the root path - shows all groups
func GroupsHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get groups data
	groups, err := database.GetGroups()
	if err != nil {
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		return
	}

	// Use the global templates variable instead of parsing the template directly
	if err := templates.ExecuteTemplate(w, "groups.html", groups); err != nil {
		log.Printf("Template execution error: %v", err)
		// Don't write header here since template.Execute might have already written it
	}
}

// GroupDetailsHandler handles group details and deletion
func GroupDetailsHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.NotFound(w, r)
		return
	}

	id := pathParts[2]
	if len(pathParts) == 4 && pathParts[3] == "delete" {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := database.DeleteGroup(database.DB, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	group, err := database.GetGroupDetails(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "group_details.html", group); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AddGroupHandler handles the addition of new groups
func AddGroupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		weapons, err := database.GetWeapons()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		vehicles, err := database.GetVehicles()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert data to JSON for the template
		weaponsJSON, err := json.Marshal(weapons)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		vehiclesJSON, err := json.Marshal(vehicles)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			Weapons        interface{}
			WeaponOptions string
			VehicleOptions string
		}{
			Weapons:        weapons,  // For the template weapon select
			WeaponOptions: string(weaponsJSON),  // For JavaScript
			VehicleOptions: string(vehiclesJSON), // For JavaScript
		}

		if err := templates.ExecuteTemplate(w, "add_group.html", data); err != nil {
			log.Printf("Template execution error: %v", err)
		}
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the country code from the hidden input
	countryCode := r.FormValue("nationality")
	if countryCode == "" {
		http.Error(w, "Invalid country code", http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert group with country code
	result, err := tx.Exec(`
		INSERT INTO groups (group_name, group_nationality, group_size)
		VALUES (?, ?, 0)
	`, r.FormValue("name"), countryCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle direct members
	roles := r.PostForm["role[]"]
	ranks := r.PostForm["rank[]"]
	totalMembers := len(roles)
	
	for i := range roles {
		// Insert member
		result, err := tx.Exec(`
			INSERT INTO members (member_role, member_rank)
			VALUES (?, ?)
		`, roles[i], ranks[i])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		memberID, err := result.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Associate member with group
		_, err = tx.Exec(`
			INSERT INTO group_members (group_id, member_id, team_id)
			VALUES (?, ?, NULL)
		`, groupID, memberID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Handle weapons for this member
		weaponIDs := r.PostForm[fmt.Sprintf("weapons_%d[]", i)]
		for _, weaponID := range weaponIDs {
			_, err = tx.Exec(`
				INSERT INTO members_weapons (member_id, weapon_id)
				VALUES (?, ?)
			`, memberID, weaponID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	// Handle teams
	teamNames := r.PostForm["team_name[]"]
	for i, name := range teamNames {
		teamRoles := r.PostForm[fmt.Sprintf("team_%d_role[]", i)]
		teamSize := len(teamRoles)
		totalMembers += teamSize

		// Insert team
		result, err := tx.Exec(`
			INSERT INTO teams (team_name, team_size)
			VALUES (?, ?)
		`, name, teamSize)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		teamID, err := result.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Associate team with group
		_, err = tx.Exec(`
			INSERT INTO group_members (group_id, member_id, team_id)
			VALUES (?, NULL, ?)
		`, groupID, teamID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Handle team members
		teamRanks := r.PostForm[fmt.Sprintf("team_%d_rank[]", i)]
		
		for j := range teamRoles {
			// Insert member
			result, err := tx.Exec(`
				INSERT INTO members (member_role, member_rank)
				VALUES (?, ?)
			`, teamRoles[j], teamRanks[j])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			memberID, err := result.LastInsertId()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Associate member with team
			_, err = tx.Exec(`
				INSERT INTO team_members (team_id, member_id)
				VALUES (?, ?)
			`, teamID, memberID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Handle weapons for this team member
			weaponIDs := r.PostForm[fmt.Sprintf("team_%d_weapons_%d[]", i, j)]
			for _, weaponID := range weaponIDs {
				_, err = tx.Exec(`
					INSERT INTO members_weapons (member_id, weapon_id)
					VALUES (?, ?)
				`, memberID, weaponID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	}

	// Handle vehicles
	vehicleIDs := r.PostForm["vehicle_id[]"]
	for i, vehicleID := range vehicleIDs {
		// Insert vehicle instance
		result, err := tx.Exec(`
			INSERT INTO group_vehicles (group_id, vehicle_id)
			VALUES (?, ?)
		`, groupID, vehicleID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		instanceID, err := result.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Handle vehicle members
		vehicleRoles := r.PostForm[fmt.Sprintf("vehicle_%d_role[]", i)]
		vehicleRanks := r.PostForm[fmt.Sprintf("vehicle_%d_rank[]", i)]
		totalMembers += len(vehicleRoles)
		
		for j := range vehicleRoles {
			// Insert member
			result, err := tx.Exec(`
				INSERT INTO members (member_role, member_rank)
				VALUES (?, ?)
			`, vehicleRoles[j], vehicleRanks[j])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			memberID, err := result.LastInsertId()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Associate member with vehicle instance
			_, err = tx.Exec(`
				INSERT INTO vehicle_members (instance_id, member_id)
				VALUES (?, ?)
			`, instanceID, memberID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Handle weapons for this vehicle member
			weaponIDs := r.PostForm[fmt.Sprintf("vehicle_%d_weapons_%d[]", i, j)]
			for _, weaponID := range weaponIDs {
				_, err = tx.Exec(`
					INSERT INTO members_weapons (member_id, weapon_id)
					VALUES (?, ?)
				`, memberID, weaponID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	}

	// Update group size
	_, err = tx.Exec("UPDATE groups SET group_size = ? WHERE group_id = ?", totalMembers, groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// EditGroupHandler handles editing existing groups
func EditGroupHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.NotFound(w, r)
		return
	}

	groupID := pathParts[2]

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get the country code from the hidden input
		countryCode := r.FormValue("nationality")
		if countryCode == "" {
			http.Error(w, "Invalid country code", http.StatusBadRequest)
			return
		}

		// Start transaction
		tx, err := database.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Update group with country code
		_, err = tx.Exec(`
			UPDATE groups 
			SET group_name = ?, group_nationality = ?
			WHERE group_id = ?
		`, r.FormValue("name"), countryCode, groupID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Process the rest of the form data...
		// ... (your existing code for processing members, teams, etc.)

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/group/%s", groupID), http.StatusSeeOther)
		return
	}

	// Handle GET request
	group, err := database.GetGroupDetails(groupID)
	if err != nil {
		log.Printf("Error getting group details: %v", err)
		http.Error(w, "Failed to get group details", http.StatusInternalServerError)
		return
	}

	// Get weapon options
	weaponOptions, err := database.GetWeapons()
	if err != nil {
		log.Printf("Error getting weapons: %v", err)
		http.Error(w, "Failed to get weapons", http.StatusInternalServerError)
		return
	}

	// Get vehicle options
	vehicleOptions, err := database.GetVehicles()
	if err != nil {
		log.Printf("Error getting vehicles: %v", err)
		http.Error(w, "Failed to get vehicles", http.StatusInternalServerError)
		return
	}

	// Convert data to JSON for template
	weaponOptionsJSON, err := json.Marshal(weaponOptions)
	if err != nil {
		log.Printf("Error marshaling weapons: %v", err)
		http.Error(w, "Failed to process weapons data", http.StatusInternalServerError)
		return
	}

	vehicleOptionsJSON, err := json.Marshal(vehicleOptions)
	if err != nil {
		log.Printf("Error marshaling vehicles: %v", err)
		http.Error(w, "Failed to process vehicles data", http.StatusInternalServerError)
		return
	}

	groupJSON, err := json.Marshal(group)
	if err != nil {
		log.Printf("Error marshaling group: %v", err)
		http.Error(w, "Failed to process group data", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Group":          string(groupJSON),
		"WeaponOptions":  string(weaponOptionsJSON),
		"VehicleOptions": string(vehicleOptionsJSON),
		"Nationality":    group.Nationality,
	}

	if err := templates.ExecuteTemplate(w, "edit_group.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		// Don't write an error header here since the template might have already written a response
	}
} 