package handlers

import (
	"encoding/json"
	"fmt"
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

	groups, err := database.GetGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "groups.html", groups); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

		data := struct {
			WeaponOptions  []interface{}
			VehicleOptions []interface{}
		}{
			WeaponOptions:  interfaceSlice(weapons),
			VehicleOptions: interfaceSlice(vehicles),
		}

		if err := templates.ExecuteTemplate(w, "add_group.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert group
	result, err := tx.Exec(`
		INSERT INTO groups (group_name, group_nationality, group_size)
		VALUES (?, ?, 0)
	`, r.FormValue("name"), r.FormValue("nationality"))
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

// EditGroupHandler handles the editing of existing groups
func EditGroupHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 || pathParts[3] != "edit" {
		http.NotFound(w, r)
		return
	}

	groupID := pathParts[2]

	if r.Method == "GET" {
		// Get group details
		group, err := database.GetGroupDetails(groupID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get weapons and vehicles for the form
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
		groupJSON, err := json.Marshal(group)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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
			Group         string
			WeaponOptions string
			VehicleOptions string
		}{
			Group:         string(groupJSON),
			WeaponOptions: string(weaponsJSON),
			VehicleOptions: string(vehiclesJSON),
		}

		if err := templates.ExecuteTemplate(w, "edit_group.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Start a transaction for the entire update process
		tx, err := database.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// First, delete the existing group and all its associated data
		err = database.DeleteGroup(tx, groupID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete existing group data: %v", err), http.StatusInternalServerError)
			return
		}

		// Insert the updated group with the same ID
		_, err = tx.Exec(`
			INSERT INTO groups (group_id, group_name, group_nationality, group_size)
			VALUES (?, ?, ?, 0)
		`, groupID, r.FormValue("name"), r.FormValue("nationality"))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert updated group: %v", err), http.StatusInternalServerError)
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

			// Handle vehicle crew members
			vehicleRoles := r.PostForm[fmt.Sprintf("vehicle_%d_role[]", i)]
			vehicleRanks := r.PostForm[fmt.Sprintf("vehicle_%d_rank[]", i)]
			totalMembers += len(vehicleRoles)

			for j := range vehicleRoles {
				// Insert crew member
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

				// Associate crew member with vehicle instance
				_, err = tx.Exec(`
					INSERT INTO vehicle_members (instance_id, member_id)
					VALUES (?, ?)
				`, instanceID, memberID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// Handle weapons for this crew member
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

		http.Redirect(w, r, fmt.Sprintf("/group/%s", groupID), http.StatusSeeOther)
	}
} 