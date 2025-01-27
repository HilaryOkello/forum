package comments

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"forum/db"
	"forum/internals/auth"
	"forum/internals/fails"
)

func ReactToComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fails.ErrorPageHandler(w, r, http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the session from the request context
	session, ok := r.Context().Value(auth.UserSessionKey).(*auth.Session)
	if !ok {
		// Handle the case where the session is not found in the context
		log.Println("Session not found in context")
		fails.ErrorPageHandler(w, r, http.StatusUnauthorized)
		return
	}

	var input reactToCommentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err.Error())
		fails.ErrorPageHandler(w, r, http.StatusBadRequest)
		return
	}

	if input.CommentID == 0 || input.ReactionType == "" {
		log.Println("Invalid input")
		fails.ErrorPageHandler(w, r, http.StatusBadRequest)
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
		log.Println(err.Error())
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}

	var responseStatus string
	if currentReaction == input.ReactionType {
		// Remove the reaction if it's the same as the current one
		queryDelete := `
            DELETE FROM comment_reactions
            WHERE comment_id = ? AND user_id = ?`
		_, err := db.DB.Exec(queryDelete, input.CommentID, session.UserID)
		if err != nil {
			log.Println(err.Error())
			fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
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
			log.Println(err.Error())
			fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
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
		log.Println(err.Error())
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}
}

// CreateComment creates a new comment
func CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fails.ErrorPageHandler(w, r, http.StatusMethodNotAllowed)
		return
	}
	// Retrieve the session from the request context
	session, ok := r.Context().Value(auth.UserSessionKey).(*auth.Session)
	if !ok {
		log.Println("Session not found in context")
		fails.ErrorPageHandler(w, r, http.StatusUnauthorized)
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

	if input.PostID == 0 || input.Content == "" {
		log.Println("Invalid input")
		fails.ErrorPageHandler(w, r, http.StatusBadRequest)
		return
	}

	var id int64
	var createdAt string

	query := `INSERT INTO comments (post_id, parent_id, content, user_id) VALUES (?, ?, ?, ?) RETURNING id, created_at`

	// Execute the query and scan the result into the id and createdAt variables
	err := db.DB.QueryRow(query, input.PostID, input.ParentID, input.Content, session.UserID).Scan(&id, &createdAt)
	if err != nil {
		log.Println(err.Error())
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}

	response := struct {
		Status    string `json:"status"`
		ID        int64  `json:"id"`
		CreatedAt string `json:"created_at"`
		Username  string `json:"username"`
	}{
		Status:    "success",
		ID:        id,
		CreatedAt: createdAt,
		Username:  session.UserName,
	}

	json.NewEncoder(w).Encode(response)
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fails.ErrorPageHandler(w, r, http.StatusMethodNotAllowed)
		return
	}

	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		fails.ErrorPageHandler(w, r, http.StatusBadRequest)
		return
	}

	// Get the current logged-in user from the session
	var userID int64
	session := auth.CheckIfLoggedIn(w, r)
	if session != nil {
		userID = int64(session.UserID)
	} else {
		userID = 0
	}

	// Get comments for the post, including the user's reactions
	comments, err := getPostComments(postID, userID)
	if err != nil {
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comments)
}
