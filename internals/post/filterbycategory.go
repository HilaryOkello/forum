package post

import (
	"fmt"
	"forum/db"  
	"strings"
	"net/http"
	"html/template"
)

// FetchPostsByCategory retrieves posts from the database filtered by category
func FetchPostsByCategory(category string) ([]Post, error) {

	category = strings.TrimSpace(category)
	category = strings.ToLower(category)

	// SQL query to fetch posts for a specific category
	query := `
		SELECT p.id, p.title, p.content, u.username
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN post_categories pc ON p.id = pc.post_id
		JOIN categories c ON pc.category_id = c.id
		WHERE LOWER(c.name) = ?
		ORDER BY p.id DESC;
	`

	// Using db.DB to query the database (db.DB is likely a *sql.DB object from the db package)
	rows, err := db.DB.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts by category: %w", err)
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

