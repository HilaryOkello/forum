package comment

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
	Children  []Comment `json:"children,omitempty"`
}
