package comments

import (
	"encoding/json"
	"fmt"
	"net/http"

	"forum/db"
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

	fmt.Println(finalRootComments)
	return finalRootComments, nil
}

// CreateComment creates a new comment
func CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input commentInput

	// Decode the request body into the input struct
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "failure",
		})
	}

	var id int64
	var createdAt string

	query := `INSERT INTO comments (post_id, parent_id, content, user_id) VALUES (?, ?, ?, ?) RETURNING id, created_at`

	// Execute the query and scan the result into the id and createdAt variables
	err := db.DB.QueryRow(query, input.PostID, input.ParentID, input.Content, input.UserID).Scan(&id, &createdAt)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string `json:"status"`
	}{Status: "success"}

	json.NewEncoder(w).Encode(response)
}

func ReactToComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input reactToCommentInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert or update the reaction for the comment
	query := `
		INSERT INTO comment_reactions (comment_id, user_id, reaction_type)
		VALUES (?, ?, ?)
		ON CONFLICT (comment_id, user_id) DO UPDATE SET reaction_type = ?`

	_, err := db.DB.Exec(query, input.CommentID, input.UserID, input.ReactionType, input.ReactionType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
