package comments

import (
	"forum/db"
)

func getPostComments(postID string, userID int64) ([]Comment, error) {
	// Query to get all comments for a post, including the user's reaction
	query := `
        SELECT 
            c.id, c.post_id, c.user_id, c.parent_id, c.content, c.created_at,
            u.username,
            (SELECT COUNT(*) FROM comment_reactions WHERE comment_id = c.id AND reaction_type = 'LIKE') as likes,
            (SELECT COUNT(*) FROM comment_reactions WHERE comment_id = c.id AND reaction_type = 'DISLIKE') as dislikes,
            (SELECT reaction_type FROM comment_reactions WHERE comment_id = c.id AND user_id = ?) as user_reaction
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.post_id = ?
        ORDER BY c.created_at DESC`

	rows, err := db.DB.Query(query, userID, postID)
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
