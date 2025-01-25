package post

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"forum/db"
	"forum/internals/auth"
)

// Post represents a post structure
type Post struct {
	ID           int
	Title        string
	Content      string
	UserName     string
	CreatedAt    time.Time
	CommentCount int
	Likes        int
	Dislikes     int
}

func FetchPosts() ([]Post, error) {
	query := `
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			u.username, 
			p.created_at,
			COALESCE(c.comment_count, 0) AS comment_count,
			COALESCE(r.likes, 0) AS likes,
			COALESCE(r.dislikes, 0) AS dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comment_count 
			FROM comments 
			GROUP BY post_id
		) c ON p.id = c.post_id
		LEFT JOIN (
			SELECT 
				post_id, 
				SUM(CASE WHEN reaction_type = 'LIKE' THEN 1 ELSE 0 END) AS likes,
				SUM(CASE WHEN reaction_type = 'DISLIKE' THEN 1 ELSE 0 END) AS dislikes
			FROM post_reactions
			GROUP BY post_id
		) r ON p.id = r.post_id
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
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserName,
			&post.CreatedAt,
			&post.CommentCount,
			&post.Likes,
			&post.Dislikes,
		)
		if err != nil {
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
	// SQL query to fetch the post with additional fields.
	query := `
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			u.username, 
			p.created_at,
			COALESCE(c.comment_count, 0) AS comment_count,
			COALESCE(r.likes, 0) AS likes,
			COALESCE(r.dislikes, 0) AS dislikes
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comment_count 
			FROM comments 
			GROUP BY post_id
		) c ON p.id = c.post_id
		LEFT JOIN (
			SELECT 
				post_id, 
				SUM(CASE WHEN reaction_type = 'LIKE' THEN 1 ELSE 0 END) AS likes,
				SUM(CASE WHEN reaction_type = 'DISLIKE' THEN 1 ELSE 0 END) AS dislikes
			FROM post_reactions
			GROUP BY post_id
		) r ON p.id = r.post_id
		WHERE p.id = ?;
	`

	// Variable to hold the fetched post.
	var post Post

	// Execute the query.
	err := db.DB.QueryRow(query, postID).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserName,
		&post.CreatedAt,
		&post.CommentCount,
		&post.Likes,
		&post.Dislikes,
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
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	post, err := fetchPostFromDB(postID) // Fetch post data from the database
	if err != nil {
		log.Println(err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Check if the user is logged in
	session := auth.CheckIfLoggedIn(w, r)

	// Create the PageData object
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
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}
