package post

import (
	"fmt"
	"forum/db"
)

// Post represents a post structure
type Post struct {
	ID       int
	Title    string
	Content  string
	UserName string
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

