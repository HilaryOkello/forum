package post

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"forum/db"
	"forum/internals/auth"
)

type PageData struct {
	IsLoggedIn bool
	UserName   string
}

func ServeCreatePostForm(w http.ResponseWriter, r *http.Request) {
	session := auth.CheckIfLoggedIn(w, r)
	var pageData PageData
	if session == nil {
		pageData = PageData{
			IsLoggedIn: false,
		}
	} else {
		pageData = PageData{
			IsLoggedIn: true,
			UserName:   session.UserName,
		}
	}

	// Parse and execute the template
	t, err := template.ParseFiles("./templates/post.html")
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, pageData); err != nil {
		log.Println(err)
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

func derefString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}

// ServeHomePage handles requests to render the homepage
func ServeHomePage(w http.ResponseWriter, r *http.Request) {
	session := auth.CheckIfLoggedIn(w, r)
	var userID int64
	var pageData PageData
	if session == nil {
		pageData = PageData{
			IsLoggedIn: false,
		}
		userID = 0
	} else {
		pageData = PageData{
			IsLoggedIn: true,
			UserName:   session.UserName,
		}
		userID = int64(session.UserID)
	}
	posts, err := FetchPosts(int64(userID))
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	t := template.Must(template.ParseFiles("./templates/index.html"))
	
	if err := t.Execute(w, map[string]interface{}{
		"Posts":    posts,
		"PageData": pageData,
	}); err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
	}
}

// ServeHomePage handles requests to return homepage data as JSON
func ServePosts(w http.ResponseWriter, r *http.Request) {
	session := auth.CheckIfLoggedIn(w, r)

	pageData := PageData{
		IsLoggedIn: true,
		UserName:   session.UserName,
	}

	posts, err := FetchPosts(int64(session.UserID))
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

func CreatePost(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the request context
	session, ok := r.Context().Value(auth.UserSessionKey).(*auth.Session) // Replace *Session with your session type
	if !ok {
		// Handle the case where the session is not found in the context
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	categoryIDs := r.Form["categories[]"]

	// Use the user ID from the session
	userID := session.UserID

	// Insert the new post into the POSTS table
	postQuery := `INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)`
	result, err := db.DB.Exec(postQuery, userID, title, content)
	if err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Get the ID of the newly created post
	postID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve post ID", http.StatusInternalServerError)
		return
	}

	// Insert into the Post_Categories table
	for _, categoryID := range categoryIDs {
		_, err := db.DB.Exec(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`, postID, categoryID)
		if err != nil {
			http.Error(w, "Failed to associate category with post", http.StatusInternalServerError)
			return
		}
	}

	// Redirect to the homepage after successful post creation
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
