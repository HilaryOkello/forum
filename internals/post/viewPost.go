package post

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"forum/db"
)

// Post represents a post structure
type Post struct {
	ID        int
	Title     string
	Content   string
	UserName  string
	CreatedAt time.Time
}

// FetchPosts retrieves posts from the database
func FetchPosts() ([]Post, error) {
	query := `
		SELECT p.id, p.title, p.content, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		ORDER BY p.id DESC;
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.UserName); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return posts, nil
}

// fetchPostFromDB retrieves a post by its ID from the database.
func fetchPostFromDB(postID string) (*Post, error) {
	// SQL query to fetch the post.
	query := `
		SELECT id, user_id, title, content, created_at
		FROM posts
		WHERE id = ?;
	`

	// Variable to hold the fetched post.
	var post Post

	// Execute the query.
	err := db.DB.QueryRow(query, postID).Scan(&post.ID, &post.Title, &post.Content, &post.UserName, &post.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post with ID %s not found", postID)
		}
		return nil, fmt.Errorf("failed to fetch post: %v", err)
	}

	return &post, nil
}

func ViewPost(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	post, err := fetchPostFromDB(postID) // Fetch post data from the database
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	fmt.Println(post)

	tmpl := template.Must(template.ParseFiles("templates/viewPost.html"))
	if err := tmpl.Execute(w, post); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		fmt.Println("Template execution error:", err)
	}
}
