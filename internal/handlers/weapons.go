package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"orbat/internal/database"
	"orbat/internal/storage"
)

// WeaponsHandler handles weapons list and weapon addition
func WeaponsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parse multipart form with 10MB max memory
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		weaponType := r.FormValue("type")
		caliber := r.FormValue("caliber")
		replace := r.FormValue("replace") == "true"
		
		// Check if weapon with this name exists
		exists, existingID, err := database.WeaponExists(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if exists && !replace {
			// Return a special status code to indicate name conflict
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Weapon with this name already exists"))
			return
		}
		
		var imageURL string
		// Handle image upload if present
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			
			filename := fmt.Sprintf("weapons/%s-%d%s", 
				strings.ToLower(strings.ReplaceAll(name, " ", "-")),
				time.Now().Unix(),
				filepath.Ext(header.Filename))
			
			uploadedURL, err := storage.UploadImage(file, filename)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			imageURL = uploadedURL
		}

		tx, err := database.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		if exists && replace {
			if imageURL != "" {
				_, err = tx.Exec(`
					UPDATE weapons 
					SET weapon_type = ?,
						weapon_caliber = ?,
						image_url = ?
					WHERE weapon_id = ?`, 
					weaponType, caliber, imageURL, existingID)
			} else {
				_, err = tx.Exec(`
					UPDATE weapons 
					SET weapon_type = ?,
						weapon_caliber = ?
					WHERE weapon_id = ?`, 
					weaponType, caliber, existingID)
			}
		} else {
			_, err = tx.Exec(`
				INSERT INTO weapons (weapon_name, weapon_type, weapon_caliber, image_url)
				VALUES (?, ?, ?, ?)`, 
				name, weaponType, caliber, imageURL)
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to commit transaction: %v", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/weapons", http.StatusSeeOther)
		return
	}

	// GET request handling
	weapons, err := database.GetWeapons()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch weapons: %v", err), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "weapons.html", weapons); err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}
}

// WeaponDetailsHandler handles weapon details and deletion
func WeaponDetailsHandler(w http.ResponseWriter, r *http.Request) {
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

		if err := database.DeleteWeapon(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/weapons", http.StatusSeeOther)
		return
	}

	details, err := database.GetWeaponDetails(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "weapon_details.html", details); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// MemberWeaponsHandler handles managing member weapons
func MemberWeaponsHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 4 || pathParts[3] != "weapons" {
		http.NotFound(w, r)
		return
	}

	memberID := pathParts[2]

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := database.UpdateMemberWeapons(memberID, r.Form["weapons[]"]); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	weapons, err := database.GetMemberWeaponsData(memberID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(weapons); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
} 