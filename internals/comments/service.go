package comments

import (
	"log"
	"net/http"

	"forum/db"
	"forum/internals/auth"
	"forum/internals/fails"
)

// getPostComments retrieves all comments for a post
func getPostComments(postID string, userID int64) ([]Comment, error) {
	// Query the database for comments
	rows, err := db.DB.Query(queryGetPostComments, userID, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create a map to store comments by ID
	commentMap := make(map[int64]*Comment)
	var rootComments []*Comment

	// First pass: load all comments into the map
	for rows.Next() {
		var comment Comment
		var ParentID *int64
		var UserReaction *string

		err := rows.Scan(
			&comment.ID, &comment.PostID, &comment.UserID, &ParentID,
			&comment.Content, &comment.CreatedAt, &comment.Username,
			&comment.Likes, &comment.Dislikes, &UserReaction,
		)
		if err != nil {
			return nil, err
		}

		// Add the parent ID and user reaction
		comment.ParentID = ParentID
		comment.UserReaction = UserReaction

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

// validateMethod checks if the request method is the expected one
func validateMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		fails.ErrorPageHandler(w, r, http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// validateSession checks if the user session is valid
func validateSession(w http.ResponseWriter, r *http.Request) (*auth.Session, bool) {
	session, ok := r.Context().Value(auth.UserSessionKey).(*auth.Session)
	if !ok {
		log.Println("Session not found in context")
		fails.ErrorPageHandler(w, r, http.StatusUnauthorized)
		return nil, false
	}
	return session, true
}
