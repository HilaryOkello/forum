package auth

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"forum/db"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

var store = NewSessionStore()

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

func encryptPassword(password string) (string, error) {
	bcryptPassword, error := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if error != nil {
		return "", error
	}
	return string(bcryptPassword), nil
}

func decryptPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// PageData represents the data structure we'll pass to our templates
type PageData struct {
	IsLoggedIn bool
	UserName   string
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		ipAddress := r.RemoteAddr
		fmt.Printf("email: %s, password: %s\n", email, password)
		users := ReadfromDb()
		// validate credentials
		for _, user := range users {
			if user.Email == email && decryptPassword(user.Password, password) {
				fmt.Printf("%s logged in successfully\n", user.UserName)
				if oldsession, ok := store.GetSessionByUserId(user.ID); ok {
					store.DeleteSession(oldsession.ID)
				}
				session := store.CreateSession(user.ID, user.UserName,  ipAddress)
				if session == nil {
					http.Error(w, "Failed to create session", http.StatusInternalServerError)
					return
				}
				// Set session cookie
				http.SetCookie(w, &http.Cookie{
					Name:     "session",
					Value:    session.ID.String(),
					Path:     "/",
					HttpOnly: true,
					MaxAge:   86400, // 24 hours
				})

				http.Redirect(w, r, "/", http.StatusSeeOther)

				return

			}
		}
		// If credentials are invalid, re-render login page with error
		if err := tmpl.ExecuteTemplate(w, "login.html", "Invalid credentials"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else if r.Method == http.MethodGet {
		if err := tmpl.ExecuteTemplate(w, "login.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func Middleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session ID from cookie
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Parse the UUID from cookie
		sessionID, err := uuid.FromString(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validate session
		_, valid := store.GetSession(sessionID)
		if !valid {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Extend session
		store.ExtendSession(sessionID)
		next.ServeHTTP(w, r)
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Get session ID from cookie
	cookie, err := r.Cookie("session")
	if err == nil {
		// Parse the UUID from cookie
		if sessionID, err := uuid.FromString(cookie.Value); err == nil {
			store.DeleteSession(sessionID)
		}
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl.ExecuteTemplate(w, "signup.html", nil)
	} else if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		name := r.FormValue("name")
		if email == "" || password == "" || name == "" {
			tmpl.ExecuteTemplate(w, "signup.html", "All fields are required")
			return
		}
		hashedPassword, err := encryptPassword(password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var user User
		user.Email = email
		user.Password = string(hashedPassword)
		user.UserName = name
		SaveUserToDb(user)

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
}

func ReadfromDb() []User {
	users := []User{}
	rows, err := db.DB.Query("SELECT id, username, email, password, created_at FROM users")
	if err != nil {
		log.Printf("Error querying users: %v", err)
		return users
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.UserName, &user.Email, &user.Password, &user.CreatedAt)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}
		users = append(users, user)
	}
	return users
}

func SaveUserToDb(user User) {
	stmt, err := db.DB.Prepare("INSERT INTO users (username, email, password, created_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return
	}
	defer stmt.Close()

	user.CreatedAt = time.Now()
	_, err = stmt.Exec(user.UserName, user.Email, user.Password, user.CreatedAt)
	if err != nil {
		log.Printf("Error saving user: %v", err)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	link := PageData{IsLoggedIn: false}
	// check if user is logged in
	cookie, err := r.Cookie("session")
	if err == nil {
		// Parse the UUID from cookie
		if sessionID, err := uuid.FromString(cookie.Value); err == nil {
			if session, valid := store.GetSession(sessionID); valid {
				link = PageData{
					IsLoggedIn: true,
					UserName:   session.UserName,
				}
				tmpl.ExecuteTemplate(w, "index.html", link)
				return
			}
		}
	}

	tmpl.ExecuteTemplate(w, "index.html", link)
}
