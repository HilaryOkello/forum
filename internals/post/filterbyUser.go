// functionality to allow a user to filter posts to only show posts they created
package post

import (
	"encoding/json"
	"fmt"
	"net/http"

	"forum/db"
	"forum/internals/auth"
)

func FilterbyUser(w http.ResponseWriter, r *http.Request) {
	// Retrieve the user name from the session
	session := auth.CheckIfLoggedIn(w, r)

	if session == nil {
		sendErrorResponse(w, "User not logged in", http.StatusUnauthorized)
		return
	}
	user := session.UserName
	fmt.Println(user)
	// Fetch the posts for the given user
	posts, err := FetchPostsByUser(user)
	
	if err != nil {
		sendErrorResponse(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	// now we send the post ids to the frontend as json
	// Extract post IDs
	postIDs := make([]int, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}

	// Set JSON content type and send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(postIDs)
}

func FetchPostsByUser(userName string) ([]Post, error) {
	var userPosts []Post
	allPosts, err := FetchPosts()
	if err != nil {
		return nil, err
	}
	// Iterate through all posts and filter by username
	for _, post := range allPosts {
		if post.UserName == userName {
			userPosts = append(userPosts, post)
		}
	}

	// Return the filtered list or an empty slice
	return userPosts, nil
}

func sendErrorResponse(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func FilterbyLikes(w http.ResponseWriter, r *http.Request) {
	// Retrieve the user id from the session
	session := auth.CheckIfLoggedIn(w, r)

	if session == nil {
		sendErrorResponse(w, "User not logged in", http.StatusUnauthorized)
		return
	}
	user := session.UserID
	fmt.Println(user)
	// Fetch the posts for the given user
	posts, err := FetchPostsByLikes(user)
	fmt.Println(posts)
	if err != nil {
		sendErrorResponse(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	
	// Set JSON content type and send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func FetchPostsByLikes(userID int) ([]int, error) {
	query := `
		SELECT post_id
		FROM post_reactions
		WHERE user_id = ? AND reaction_type = 'LIKE';
	`

	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch liked post IDs: %w", err)
	}
	defer rows.Close()

	var postIDs []int
	for rows.Next() {
		var postID int
		if err := rows.Scan(&postID); err != nil {
			return nil, fmt.Errorf("failed to scan post ID: %w", err)
		}
		postIDs = append(postIDs, postID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return postIDs, nil
}
