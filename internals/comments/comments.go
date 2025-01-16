package comments

import (
	"encoding/json"
	"forum/db"
	"net/http"
)

func GetComments(w http.ResponseWriter, r *http.Request) {
	

}

func GetCommentsForPost(w http.ResponseWriter, r *http.Request) {

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
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	var id int64
	var createdAt string

	query := `INSERT INTO comments (post_id, parent_id, content, user_id) VALUES ($?, $?, $?, $?) RETURNING id, createdAt`

	// Execute the query and scan the result into the id and createdAt variables
	err := db.DB.QueryRow(query, input.PostID, input.ParentID, input.Content, input.UserID).Scan(&id, &createdAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		ID  int64  `json:"id"`
		CreatedAt string `json:"created_at"`
	}{id, createdAt}

	json.NewEncoder(w).Encode(response)
}

func ReactToComment(w http.ResponseWriter, r *http.Request) {
}
