package post

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"forum/db"
	"forum/internals/auth"
	"forum/internals/fails"
)

func ReactToPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fails.ErrorPageHandler(w, r, http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the session from the request context
	session, ok := r.Context().Value(auth.UserSessionKey).(*auth.Session)

	if !ok {
		// Handle the case where the session is not found in the context
		fails.ErrorPageHandler(w, r, http.StatusUnauthorized)
		return
	}

	var input reactToPost
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Printf("check one: %v ", err)
		fails.ErrorPageHandler(w, r, http.StatusBadRequest)
		return
	}

	if input.PostID <= 0 {
		fails.ErrorPageHandler(w, r, http.StatusBadRequest)
		return
	}

	// Check the current reaction for the user and post.
	var currentReaction string
	queryCheck := `
        SELECT reaction_type
        FROM post_reactions
        WHERE post_id = ? AND user_id = ?`

	err := db.DB.QueryRow(queryCheck, input.PostID, session.UserID).Scan(&currentReaction)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err.Error())
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}

	var responseStatus string
	if currentReaction == input.ReactionType {
		// Remove the reaction if it's the same as the current one
		queryDelete := `
            DELETE FROM post_reactions
            WHERE post_id = ? AND user_id = ?`
		_, err := db.DB.Exec(queryDelete, input.PostID, session.UserID)
		if err != nil {
			fmt.Println(err.Error())
			fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
			return
		}
		responseStatus = "removed"
	} else {
		// Insert or update the reaction
		queryUpsert := `
            INSERT INTO post_reactions (post_id, user_id, reaction_type)
            VALUES (?, ?, ?)
            ON CONFLICT (post_id, user_id) DO UPDATE SET reaction_type = ?`
		_, err := db.DB.Exec(queryUpsert, input.PostID, session.UserID, input.ReactionType, input.ReactionType)
		if err != nil {
			fmt.Println(err.Error())
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
		fmt.Println(err.Error())
		fails.ErrorPageHandler(w, r, http.StatusInternalServerError)
		return
	}
}
