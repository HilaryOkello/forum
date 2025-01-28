package post

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"forum/db"
	"forum/internals/auth"
	"forum/internals/fails"
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
	UserReaction string `json:"user_reaction,omitempty"`
}

func FetchPosts(userID int64) ([]Post, error) {
	query := `
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			u.username, 
			p.created_at,
			COALESCE(c.comment_count, 0) AS comment_count,
			COALESCE(r.likes, 0) AS likes,
			COALESCE(r.dislikes, 0) AS dislikes,
			COALESCE(pr.reaction_type, '') AS user_reaction -- Use COALESCE to handle NULL
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
		LEFT JOIN (
			SELECT post_id, reaction_type
			FROM post_reactions
			WHERE user_id = ?
		) pr ON p.id = pr.post_id
		ORDER BY p.id DESC;
	`

	rows, err := db.DB.Query(query, userID)
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
			&post.UserReaction, // Directly scan into a string
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
func fetchPostFromDB(postID string, userID int64) (*Post, error) {
	// SQL query to fetch the post with additional fields, including the user's reaction.
	query := `
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			u.username, 
			p.created_at,
			COALESCE(c.comment_count, 0) AS comment_count,
			COALESCE(r.likes, 0) AS likes,
			COALESCE(r.dislikes, 0) AS dislikes,
			COALESCE(pr.reaction_type, '') AS user_reaction -- Fetch user's reaction or default to empty string
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
		LEFT JOIN (
			SELECT post_id, reaction_type
			FROM post_reactions
			WHERE user_id = ?
		) pr ON p.id = pr.post_id
		WHERE p.id = ?;
	`

	// Variable to hold the fetched post.
	var post Post

	// Execute the query.
	err := db.DB.QueryRow(query, userID, postID).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
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
			fails.ErrorPageHandler(w,r, http.StatusInternalServerError)
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
