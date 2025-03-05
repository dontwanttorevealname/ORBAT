package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"orbat/internal/database"
	"log"
	"encoding/json"
	"github.com/biter777/countries"
)

// CountriesHandler handles the countries list
func CountriesHandler(w http.ResponseWriter, r *http.Request) {
	// Get countries data
	countries, err := database.GetCountries()
	if err != nil {
		http.Error(w, "Failed to fetch countries", http.StatusInternalServerError)
		return
	}

	// Use the global templates variable instead of creating a new one
	if err := templates.ExecuteTemplate(w, "countries.html", countries); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// CountryDetailsHandler handles country details and editing
func CountryDetailsHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.NotFound(w, r)
		return
	}

	countryName := pathParts[2]
	
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newName := r.FormValue("name")
		if newName == "" {
			http.Error(w, "Country name cannot be empty", http.StatusBadRequest)
			return
		}

		// Validate and get the standardized country code
		country := countries.ByName(newName)
		if country == countries.Unknown {
			http.Error(w, "Invalid country name", http.StatusBadRequest)
			return
		}
		newCode := country.Info().Alpha2

		tx, err := database.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Update country code in groups table
		_, err = tx.Exec("UPDATE groups SET group_nationality = ? WHERE group_nationality = ?", 
			newCode, countryName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/country/"+url.PathEscape(country.Info().Name), http.StatusSeeOther)
		return
	}

	details, err := database.GetCountryDetails(countryName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "country_details.html", details); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ValidateCountryHandler handles country validation
func ValidateCountryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	countryName := r.URL.Query().Get("name")
	if countryName == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
		})
		return
	}

	// Try to find the country
	country := countries.ByName(countryName)
	if country == countries.Unknown {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
		})
		return
	}

	info := country.Info()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid": true,
		"standardName": info.Name,
		"code": info.Alpha2,
	})
} 