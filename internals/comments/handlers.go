package comments

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"forum/db"
	"forum/internals/auth"
	"forum/internals/fails"
)

// ReactToComment handles the reaction to a comment
func ReactToComment(w http.ResponseWriter, r *http.Request) {
	if !validateMethod(w, r, http.MethodPost) {
		return
	}

	// Retrieve the session from the request context
	session, ok := validateSession(w, r)
	if !ok {
		return // validateSession already handles the response
	}

	// Decode the request body into the input struct
	var input reactToCommentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err.Error())
		fails.ErrorPageHandler(w, r, http.StatusBadRequest)
		return
	}

	// Check if the input is valid
	if input.CommentID == 0 {
		log.Println("Invalid input")
		fails.ErrorPageHandler(w, r, http.StatusBadRequest)
		return
	}

	// Check if the reaction type is valid
	validReactions := map[string]bool{"LIKE": true, "DISLIKE": true}
	if !validReactions[input.ReactionType] {
		log.Println("Invalid reaction type")
		fails.ErrorPageHandler(w, r, http.StatusBadRequest)
		return
	}

	// Check the current reaction for the user and comment
	var currentReaction string
	err := db.DB.QueryRow(QueryCheckReaction, input.CommentID, session.UserID).Scan(&currentReaction)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err.Error())
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}

	var responseStatus string

	// If the current reaction is the same as the input, remove the reaction
	if currentReaction == input.ReactionType {
		_, err := db.DB.Exec(QueryDeleteReaction, input.CommentID, session.UserID)
		if err != nil {
			log.Println(err.Error())
			fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
			return
		}
		responseStatus = "removed"
	} else {
		// Insert or update the reaction
		_, err := db.DB.Exec(QueryUpsertReaction, input.CommentID, session.UserID, input.ReactionType, input.ReactionType)
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
	if !validateMethod(w, r, http.MethodPost) {
		return
	}

	// Retrieve the session from the request context
	session, ok := validateSession(w, r)
	if !ok {
		return // validateSession already handles the response
	}

	// Decode the request body into the input struct
	var input commentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failure",
			"message": "Invalid JSON format",
		})
		return
	}

	// Check if the input is valid
	if input.PostID == 0 || input.Content == "" {
		log.Println("Invalid post ID or empty content")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failure",
			"message": "Invalid post ID or empty content",
		})
		return
	}

	var id int64
	var createdAt string

	// Execute the query and scan the result into the id and createdAt variables
	err := db.DB.QueryRow(QueryCreateComment, input.PostID, input.ParentID, input.Content, session.UserID).Scan(&id, &createdAt)
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Failed to encode JSON response:", err)
		http.Error(w, `{"status": "failure", "message": "Internal server error"}`, http.StatusInternalServerError)
	}
}

// GetComments retrieves comments for a post
func GetComments(w http.ResponseWriter, r *http.Request) {
	if !validateMethod(w, r, http.MethodGet) {
		return
	}

	// Validate and parse post_id
	postIDStr := r.URL.Query().Get("post_id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil || postID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failure",
			"message": "Invalid post_id",
		})
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

	// Fetch comments from the database
	comments, err := getPostComments(postIDStr, userID)
	if err != nil {
		log.Println("Error fetching comments:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failure",
			"message": "Failed to retrieve comments",
		})
		return
	}

	// Encode the comments as JSON and send the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(comments); err != nil {
		log.Println("Failed to encode JSON response:", err)
		http.Error(w, `{"status": "failure", "message": "Internal server error"}`, http.StatusInternalServerError)
	}
}
