package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"orbat/internal/database"
	"orbat/internal/storage"
)

// VehiclesHandler handles vehicles list and vehicle addition
func VehiclesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")
		vehicleType := r.FormValue("type")
		armament := r.FormValue("armament")
		if armament == "" {
			armament = "None"
		}

		// Check for duplicate names
		var exists bool
		err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM vehicles WHERE vehicle_name = ?)", name).Scan(&exists)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if exists && r.FormValue("replace") != "true" {
			w.WriteHeader(http.StatusConflict)
			return
		}

		tx, err := database.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var vehicleID int64
		if exists {
			// Update existing vehicle
			_, err = tx.Exec(`
				UPDATE vehicles 
				SET vehicle_type = ?, vehicle_armament = ?
				WHERE vehicle_name = ?`,
				vehicleType, armament, name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = tx.QueryRow("SELECT vehicle_id FROM vehicles WHERE vehicle_name = ?", name).Scan(&vehicleID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// Insert new vehicle
			result, err := tx.Exec(`
				INSERT INTO vehicles (vehicle_name, vehicle_type, vehicle_armament)
				VALUES (?, ?, ?)`,
				name, vehicleType, armament)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			vehicleID, err = result.LastInsertId()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Handle image upload
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			// Upload image to GCS
			filename := fmt.Sprintf("vehicles/%d_%s", vehicleID, header.Filename)
			imageURL, err := storage.UploadImage(file, filename)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Update vehicle with image URL
			_, err = tx.Exec("UPDATE vehicles SET image_url = ? WHERE vehicle_id = ?", imageURL, vehicleID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/vehicles", http.StatusSeeOther)
		return
	}

	vehicles, err := database.GetVehicles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "vehicles.html", vehicles); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// VehicleDetailsHandler handles vehicle details and deletion
func VehicleDetailsHandler(w http.ResponseWriter, r *http.Request) {
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

		if err := database.DeleteVehicle(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/vehicles", http.StatusSeeOther)
		return
	}

	details, err := database.GetVehicleDetails(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "vehicle_details.html", details); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
} 