package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"forum/db"
	"forum/internals/post"

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
				session := store.CreateSession(user.ID, user.UserName, ipAddress)
				if session == nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create session"})
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

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{
					"status":   "success",
					"username": user.UserName,
				})
				return
			}
		}
		// If credentials are invalid, send error JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
		return

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
        name := r.FormValue("username") // Changed from "name" to match the form
        
        if email == "" || password == "" || name == "" {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "All fields are required"})
            return
        }

        hashedPassword, err := encryptPassword(password)
        if err != nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
            return
        }

        var user User
        user.Email = email
        user.Password = string(hashedPassword)
        user.UserName = name

        err = SaveUserToDb(user)
        if err != nil {
            // If there's an error (likely user already exists)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusConflict)
            json.NewEncoder(w).Encode(map[string]string{"error": "User already exists"})
            return
        }

        // Success case
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "success",
            "username": user.UserName,
        })
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

func SaveUserToDb(user User)error {
	stmt, err := db.DB.Prepare("INSERT INTO users (username, email, password, created_at) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer stmt.Close()

	user.CreatedAt = time.Now()
	_, err = stmt.Exec(user.UserName, user.Email, user.Password, user.CreatedAt)
	if err!=nil{
		return err
	}
	return nil
}

// ServeHomePage handles requests to render the homepage
func ServeHomePage(w http.ResponseWriter, r *http.Request) {
	posts, err := post.FetchPosts()
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}
	t, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, map[string]interface{}{
		"Posts": posts,
	}); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}

}
