package comments

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"forum/db"
	"forum/internals/auth"
)

// GetComments returns all comments for a post
func GetComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Error(w, "post_id is required", http.StatusBadRequest)
		return
	}

	comments, err := getCommentsForPost(postID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comments)
}

// getCommentsForPost returns all comments for a post
func getCommentsForPost(postID string) ([]Comment, error) {
	// Query to get all comments for a post
	query := `
        SELECT 
            c.id, c.post_id, c.user_id, c.parent_id, c.content, c.created_at,
            u.username,
            (SELECT COUNT(*) FROM comment_reactions WHERE comment_id = c.id AND reaction_type = 'LIKE') as likes,
            (SELECT COUNT(*) FROM comment_reactions WHERE comment_id = c.id AND reaction_type = 'DISLIKE') as dislikes
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at DESC`

	rows, err := db.DB.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	commentMap := make(map[int64]*Comment)
	var rootComments []*Comment

	// First pass: load all comments into the map
	for rows.Next() {
		var comment Comment
		var ParentID *int64

		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &ParentID,
			&comment.Content, &comment.CreatedAt, &comment.Username,
			&comment.Likes, &comment.Dislikes,
		)
		if err != nil {
			return nil, err
		}

		// Add the parent ID
		comment.ParentID = ParentID

		// Add the comment to the map and a temporary list
		commentMap[comment.ID] = &comment

		// If the comment has no parent, it's a root comment
		if ParentID == nil {
			rootComments = append(rootComments, &comment)
		}
	}

	// Second pass: build the hierarchy
	for _, comment := range commentMap {
		if comment.ParentID != nil {
			parent := commentMap[*comment.ParentID]
			if parent != nil {
				parent.Children = append(parent.Children, comment)
			}
		}
	}

	// Convert root comments from []*Comment to []Comment for the return type
	finalRootComments := make([]Comment, len(rootComments))
	for i, root := range rootComments {
		finalRootComments[i] = *root
	}

	return finalRootComments, nil
}

// CreateComment creates a new comment
func CreateComment(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the request context
	session, ok := r.Context().Value(auth.UserSessionKey).(*auth.Session)
	if !ok {
		// Handle the case where the session is not found in the context
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input commentInput

	// Decode the request body into the input struct
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "failure",
		})
	}

	var id int64
	var createdAt string

	query := `INSERT INTO comments (post_id, parent_id, content, user_id) VALUES (?, ?, ?, ?) RETURNING id, created_at`

	// Execute the query and scan the result into the id and createdAt variables
	err := db.DB.QueryRow(query, input.PostID, input.ParentID, input.Content, session.UserID).Scan(&id, &createdAt)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status    string `json:"status"`
		ID        int64  `json:"id"`
		CreatedAt string `json:"created_at"`
	}{
		Status:    "success",
		ID:        id,
		CreatedAt: createdAt,
	}

	json.NewEncoder(w).Encode(response)
}

func ReactToComment(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session from the request context
	session, ok := r.Context().Value(auth.UserSessionKey).(*auth.Session)
	if !ok {
		// Handle the case where the session is not found in the context
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input reactToCommentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check the current reaction for the user and comment
	var currentReaction string
	queryCheck := `
        SELECT reaction_type
        FROM comment_reactions
        WHERE comment_id = ? AND user_id = ?`

	err := db.DB.QueryRow(queryCheck, input.CommentID, session.UserID).Scan(&currentReaction)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var responseStatus string
	if currentReaction == input.ReactionType {
		// Remove the reaction if it's the same as the current one
		queryDelete := `
            DELETE FROM comment_reactions
            WHERE comment_id = ? AND user_id = ?`
		_, err := db.DB.Exec(queryDelete, input.CommentID, session.ID)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		responseStatus = "removed"
	} else {
		// Insert or update the reaction
		queryUpsert := `
            INSERT INTO comment_reactions (comment_id, user_id, reaction_type)
            VALUES (?, ?, ?)
            ON CONFLICT (comment_id, user_id) DO UPDATE SET reaction_type = ?`
		_, err := db.DB.Exec(queryUpsert, input.CommentID, session.UserID, input.ReactionType, input.ReactionType)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if currentReaction == "" {
			responseStatus = "added"
		} else {
			responseStatus = "updated"
		}
	}

	// Send the response back to the client
	response := map[string]string{
		"status":           responseStatus,
		"updatedReaction":  input.ReactionType,
		"previousReaction": currentReaction,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
