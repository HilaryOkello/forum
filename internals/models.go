package internals

import "time"

type User struct {
	ID        int
	UserName  string
	Email     string
	Password  string
	CreatedAt time.Time
}
