package auth

import "time"

type User struct {
	ID        int
	UserName  string
	Email     string
	Password  string
	CreatedAt time.Time
}

var store = NewSessionStore()

type contextKey string

const UserSessionKey contextKey = "userSession"

// PageData represents the data structure we'll pass to our templates
type PageData struct {
	IsLoggedIn bool
	UserName   string
}