package post

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"forum/db"
	"forum/internals/auth"
)

type PageData struct {
	IsLoggedIn bool
	UserName   string
}

func ServeCreatePostForm(w http.ResponseWriter, r *http.Request) {
	username := auth.CheckIfLoggedIn(w, r)

	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pageData := PageData{
		IsLoggedIn: true,
		UserName:   username,
	}

	// Parse and execute the template
	t, err := template.ParseFiles("./templates/post.html")
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, pageData); err != nil {
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

// ServeHomePage handles requests to render the homepage
func ServeHomePage(w http.ResponseWriter, r *http.Request) {
	
username := auth.CheckIfLoggedIn(w, r)
var pageData PageData
if username == "" {
	pageData = PageData{
		IsLoggedIn: false,
		
	}
}else{
	pageData = PageData{
		IsLoggedIn: true,
		UserName:   username,
	}
	}

	posts, err := FetchPosts()
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	t, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, map[string]interface{}{
		"Posts": posts,
		"PageData":  pageData,
	}); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// ServeHomePage handles requests to return homepage data as JSON
func ServePosts(w http.ResponseWriter, r *http.Request) {
	username := auth.CheckIfLoggedIn(w, r)
	var pageData PageData
	if username == "" {
		pageData = PageData{
			IsLoggedIn: false,
		}
	} else {
		pageData = PageData{
			IsLoggedIn: true,
			UserName:   username,
		}
	}

	posts, err := FetchPosts()
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Set response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Create response structure
	response := map[string]interface{}{
		"posts":    posts,
		"pageData": pageData,
	}

	// Encode and send JSON response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}