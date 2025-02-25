package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"reflect"
	
	"orbat/internal/database"
)

// Templates is the global template cache
var templates *template.Template

// Initialize sets up the templates
func Initialize(templatesDir string) error {
	var err error
	templates, err = template.ParseGlob(filepath.Join(templatesDir, "*.html"))
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