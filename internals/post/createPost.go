package post

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"forum/db"
)

func ServeCreatePostForm(w http.ResponseWriter, r *http.Request) {
	// Parse and execute the template
	t, err := template.ParseFiles("./templates/post.html")
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
	}
}

// Category represents a single category
type Category struct {
	ID          int
	Name        string
	Description string
}

// FetchCategories retrieves categories with descriptions from the database
func FetchCategories() ([]Category, error) {
	rows, err := db.DB.Query("SELECT id, name, description FROM categories")
	// rows, err := db.DB
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var categories []Category

	// Loop through rows and scan data into the categories slice
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name, &category.Description); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		categories = append(categories, category)
	}

	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return categories, nil
}

// ServeCategories is the HTTP handler to serve categories as JSON
func ServeCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := FetchCategories()
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode the categories into JSON and send it as a response
	if err := json.NewEncoder(w).Encode(categories); err != nil {
		http.Error(w, "Failed to encode categories", http.StatusInternalServerError)
	}
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
}
