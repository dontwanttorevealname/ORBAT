package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"orbat/internal/database"
)

// CountriesHandler handles the countries list
func CountriesHandler(w http.ResponseWriter, r *http.Request) {
	countries, err := database.GetCountries()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "countries.html", countries); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

		tx, err := database.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Update country name in groups table
		_, err = tx.Exec("UPDATE groups SET group_nationality = ? WHERE group_nationality = ?", 
			newName, countryName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/country/"+url.PathEscape(newName), http.StatusSeeOther)
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