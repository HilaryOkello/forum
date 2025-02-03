package post

import "time"

// Post represents a post structure
type Post struct {
	ID           int
	Title        string
	Content      string
	Image        *string
	UserName     string
	CreatedAt    time.Time
	CommentCount int
	Likes        int
	Dislikes     int
	UserReaction string `json:"user_reaction,omitempty"`
}

type reactToPost struct {
	PostID       int64  `json:"post_id"`
	UserID       int64  `json:"user_id"`
	ReactionType string `json:"reaction_type"`
}

// Category represents a single category
type Category struct {
	ID          int
	Name        string
	Description string
}

type PageData struct {
	IsLoggedIn bool
	UserName   string
}

type ImageUploadResult struct {
	Filename string
	Error    error
	// mutex    sync.Mutex
}
