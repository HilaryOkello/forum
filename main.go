package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/db"
	"forum/internals/auth"
	"forum/internals/comments"
	"forum/internals/post"
)

func main() {
	// Initialize the database
	err := db.Initialize()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	// public routes
	mux.HandleFunc("/posts", post.ServePosts)
	mux.HandleFunc("/", post.ServeHomePage)


	mux.HandleFunc("/login", auth.Login)
	mux.HandleFunc("/logout", auth.Logout)
	mux.HandleFunc("/signup", auth.Signup)
	mux.HandleFunc("/create-post-form", post.ServeCreatePostForm)
	mux.HandleFunc("/categories", post.ServeCategories)
	mux.HandleFunc("/create-post", auth.CreatePost)
	mux.HandleFunc("/view-post", post.ViewPost)
	mux.HandleFunc("/category", post.ViewPostsByCategory)

	// Comment Routes
	mux.HandleFunc("/comments", comments.GetComments)
	mux.HandleFunc("/comments/create", comments.CreateComment)
	mux.HandleFunc("/comments/react", comments.ReactToComment)
	// static
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// // private routes
	// protected := http.NewServeMux()
	// protected.HandleFunc("/dashboard", auth.Dashboard)

	// mux.Handle("/dashboard", auth.Middleware(protected))

	fmt.Println("Server running http://localhost:8080/  and go to /login to login")
	http.ListenAndServe(":8080", mux)
}
