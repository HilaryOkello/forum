package post

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"forum/db"
	"forum/internals/auth"
	"forum/internals/fails"
)

func FetchPosts(userID int64) ([]Post, error) {

	rows, err := db.DB.Query(FetchAllPostsWithMetadata, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var imgPtr *string // Temporary variable to handle NULL image values

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&imgPtr, // Scan into the temporary image pointer
			&post.UserName,
			&post.CreatedAt,
			&post.CommentCount,
			&post.Likes,
			&post.Dislikes,
			&post.UserReaction,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}

		// Only set the image if it's not NULL
		if imgPtr != nil {
			post.Image = imgPtr
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating post rows: %w", err)
	}

	// // Debug logging - remove in production
	// for _, post := range posts {
	// 	fmt.Printf("Fetched post: ID=%d, Title=%s, Image=%v\n",
	// 		post.ID, post.Title, post.Image)
	// }

	return posts, nil
}

// fetchPostFromDB retrieves a post by its ID from the database.
func fetchPostFromDB(postID string, userID int64) (*Post, error) {

	// Variable to hold the fetched post.
	var post Post

	// Execute the query.
	err := db.DB.QueryRow(FetchPostWithUserReaction, userID, postID).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Image,
		&post.UserName,
		&post.CreatedAt,
		&post.CommentCount,
		&post.Likes,
		&post.Dislikes,
		&post.UserReaction, // Populate the UserReaction field
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post with ID %s not found", postID)
		}
		return nil, fmt.Errorf("failed to fetch post: %v", err)
	}

	return &post, nil
}

func ViewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fails.ErrorPageHandler(w, r, http.StatusMethodNotAllowed)
		return
	}

	postID := r.URL.Query().Get("id")
	if postID == "" {
		fails.ErrorPageHandler(w, r, http.StatusBadRequest)
		return
	}

	// Check if the user is logged in
	session := auth.CheckIfLoggedIn(w, r)

	// Create the PageData object
	var pageData PageData

	var userID int64

	if session == nil {
		pageData = PageData{
			IsLoggedIn: false,
		}
		userID = 0
	} else {
		// Validate session data
		if session.UserID <= 0 || session.UserName == "" {
			fails.ErrorPageHandler(w, r, http.StatusUnauthorized)
			return
		}

		pageData = PageData{
			IsLoggedIn: true,
			UserName:   session.UserName,
		}
		userID = int64(session.UserID)
	}

	post, err := fetchPostFromDB(postID, userID) // Fetch post data from the database
	if err != nil {
		log.Println(err)
		switch {
		case strings.Contains(err.Error(), "not found"):
			fails.ErrorPageHandler(w, r, http.StatusNotFound)
		case strings.Contains(err.Error(), "invaid post ID"):
			fails.ErrorPageHandler(w, r, http.StatusBadRequest)
		default:
			fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		}
		return
	}

	if post == nil || post.ID == 0 {
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}

	response := struct {
		PageData
		Post *Post
	}{
		PageData: pageData,
		Post:     post,
	}

	tmpl := template.Must(template.ParseFiles("templates/viewPost.html"))
	if err := tmpl.Execute(w, response); err != nil {
		log.Println("Template execution error:", err)
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
	}
}

func ServeAboutPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fails.ErrorPageHandler(w, r, http.StatusNotFound)
		return
	}
	
	if r.URL.Path != "/about" {
		fails.ErrorPageHandler(w, r, http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		fails.ErrorPageHandler(w, r, http.StatusMethodNotAllowed)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/about.html"))
	if err := tmpl.Execute(w, nil); err != nil {
		log.Println("Template execution error:", err)
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
	}
}
