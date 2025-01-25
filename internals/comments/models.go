package comments

// Comment represents a single comment
type Comment struct {
	ID        int64     `json:"id"`
	PostID    int64     `json:"post_id"`
	UserID    int64     `json:"user_id"`
	ParentID  *int64    `json:"parent_id,omitempty"`
	Content   string    `json:"content"`
	CreatedAt string    `json:"created_at"`
	Username  string    `json:"username"`
	Likes     int       `json:"likes"`
	Dislikes  int       `json:"dislikes"`
	UserReaction *string    `json:"user_reaction,omitempty"`
	Children  []*Comment `json:"children,omitempty"`
}

// CommentInput represents the input for creating a comment
type commentInput struct { 
    PostID   int64  `json:"post_id"`
    ParentID *int64 `json:"parent_id"`
    Content  string `json:"content"`
    UserID   int64  `json:"user_id"`
}

type reactToCommentInput struct {
	CommentID    int64  `json:"comment_id"`
	UserID       int64  `json:"user_id"`
	ReactionType string `json:"reaction_type"`
}
