package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"reflect"
	"fmt"
	"strings"
	
	"orbat/internal/database"
	"github.com/biter777/countries"
)

// Templates is the global template cache
var templates *template.Template

// Initialize sets up the templates with custom functions
func Initialize(templatesDir string) error {
	var err error
	
	// Create function map
	funcMap := template.FuncMap{
		"countryCode": func(name string) string {
			country := countries.ByName(name)
			if country != countries.Unknown {
				return country.Info().Alpha2
			}
			return name // Fallback to original name if not found
		},
		"countryFlag": func(name string) template.HTML {
			country := countries.ByName(name)
			if country != countries.Unknown {
				code := strings.ToLower(country.Info().Alpha2)
				// Return the country flag using Bootstrap's flag icons
				return template.HTML(fmt.Sprintf(`<i class="fi fi-%s"></i>`, code))
			}
			return template.HTML(`<i class="bi bi-flag"></i>`) // Fallback to generic flag
		},
	}
	
	// Parse templates with the function map
	templates, err = template.New("").Funcs(funcMap).ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		return err
	}
	return nil
}

// HealthCheckHandler handles the health check endpoint
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := database.DB.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database connection error: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Helper function to convert a slice to a slice of interfaces
func interfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return nil
	}

	ret := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
} 