package post

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"forum/db"
	"forum/internals/auth"
)

// FetchPostsByCategory retrieves posts from the database filtered by category
func FetchPostsByCategory(category string) ([]Post, error) {
	category = strings.TrimSpace(category)
	category = strings.ToLower(category)

	// SQL query to fetch posts for a specific category with additional fields
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
		JOIN post_categories pc ON p.id = pc.post_id
		JOIN categories c ON pc.category_id = c.id
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
		WHERE LOWER(c.name) = ?
		ORDER BY p.id DESC;
	`

	// Using db.DB to query the database
	rows, err := db.DB.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts by category: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserName,
			&post.CreatedAt,
			&post.CommentCount,
			&post.Likes,
			&post.Dislikes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
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
		http.Error(w, "Category is required", http.StatusBadRequest)
		return
	}

	// Fetch the posts for the given category
	posts, err := FetchPostsByCategory(category)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
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
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		fmt.Println("Template execution error:", err)
	}
}
