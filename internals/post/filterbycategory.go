package post

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"forum/db"
	"forum/internals/auth"
	"forum/internals/fails"
)

// FetchPostsByCategory retrieves posts from the database filtered by category
func FetchPostsByCategory(category string, userID int64) ([]Post, error) {
	category = strings.TrimSpace(category)
	category = strings.ToLower(category)

	rows, err := db.DB.Query(GetFilteredPostsByCategory, userID, category)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts by category: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var imgPtr *string // Temporary variable to handle NULL image values

		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&imgPtr, // Scan the image column
			&post.UserName,
			&post.CreatedAt,
			&post.CommentCount,
			&post.Likes,
			&post.Dislikes,
			&post.UserReaction,
		); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		// Only set the image if it's not NULL
		if imgPtr != nil {
			post.Image = imgPtr
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return posts, nil
}

// ViewPostsByCategory filters posts based on category and renders the filtered posts in a new template.
func ViewPostsByCategory(w http.ResponseWriter, r *http.Request) {
	// Retrieve the category from the query parameter
	category := r.URL.Query().Get("name")

	if category == "" {
		fails.ErrorPageHandler(w, r, http.StatusBadRequest)

		return
	}

	// Check if the user is logged in
	session := auth.CheckIfLoggedIn(w, r)

	var userID int64
	// Create the PageData object
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

	// Fetch the posts for the given category
	posts, err := FetchPostsByCategory(category, userID)
	if err != nil {
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}

	// Prepare the data to be passed to the template
	data := struct {
		Category string
		Posts    []Post
		PageData PageData
	}{
		Category: category,
		Posts:    posts,
		PageData: pageData,
	}

	// Parse and execute the filteredPosts.html template
	tmpl := template.Must(template.ParseFiles("templates/filterposts.html"))
	if err := tmpl.Execute(w, data); err != nil {
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		fmt.Println("Template execution error:", err)
	}
}
