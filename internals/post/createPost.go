package post

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"forum/db"
	"forum/internals/auth"
	"forum/internals/fails"
)

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
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, pageData); err != nil {
		log.Println(err)
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
	}
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
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Set the response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode the categories into JSON and send it as a response
	if err := json.NewEncoder(w).Encode(categories); err != nil {
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
	}
}

// ServeHomePage handles requests to render the homepage
func ServeHomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		fails.ErrorPageHandler(w, r, http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		fails.ErrorPageHandler(w, r, http.StatusMethodNotAllowed)
		return
	}

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
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}
	t := template.Must(template.ParseFiles("./templates/index.html"))

	if err := t.Execute(w, map[string]interface{}{
		"Posts":    posts,
		"PageData": pageData,
	}); err != nil {
		fmt.Println(err)
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
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
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
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
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	session, ok := r.Context().Value(auth.UserSessionKey).(*auth.Session)
	if !ok || session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	categoryIDs := r.Form["categories[]"]

	// Validate input
	if err := validatePostInput(title, content, categoryIDs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Begin transaction
	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Get the image filename if one was uploaded
	var imageFilename string
	uploadMutex.Lock()
	if upload, exists := currentUpload[int64(session.UserID)]; exists {
		imageFilename = upload.Filename
		fmt.Println("Image filename:", imageFilename)
		delete(currentUpload, int64(session.UserID))
	}
	uploadMutex.Unlock()

	// Insert post
	postID, err := insertPost(tx, session.UserID, title, content, imageFilename)
	if err != nil {
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	// Insert categories
	if err := insertPostCategories(tx, postID, categoryIDs); err != nil {
		http.Error(w, "Error assigning categories", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Error saving post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func validatePostInput(title, content string, categoryIDs []string) error {
	if title == "" || content == "" || len(categoryIDs) == 0 {
		return fmt.Errorf("missing required fields")
	}
	if len(title) > 50 {
		return fmt.Errorf("title exceeds 50 characters")
	}
	if len(content) < 5 {
		return fmt.Errorf("content is too short")
	}
	if len(content) > 500 {
		return fmt.Errorf("content exceeds limit")
	}
	return nil
}

func insertPost(tx *sql.Tx, userID int, title, content, imageFilename string) (int64, error) {
	var result sql.Result
	var err error

	if imageFilename != "" {
		result, err = tx.Exec(
			`INSERT INTO posts (user_id, title, content, image) VALUES (?, ?, ?, ?)`,
			userID, title, content, imageFilename,
		)
	} else {
		result, err = tx.Exec(
			`INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)`,
			userID, title, content,
		)
	}

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func insertPostCategories(tx *sql.Tx, postID int64, categoryIDs []string) error {
	for _, categoryID := range categoryIDs {
		_, err := tx.Exec(
			`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`,
			postID, categoryID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
