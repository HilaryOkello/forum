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
// post.UpdatePostImage(9,"static/images/QVC5ZG24TBI3HHTKBFXU2IEYPQ.jpg")
	mux := http.NewServeMux()
	// public routes
	mux.HandleFunc("/posts", post.ServePosts)
	mux.HandleFunc("/", post.ServeHomePage)
	mux.HandleFunc("/about", post.ServeAboutPage)

	mux.HandleFunc("/login", auth.Login)
	mux.HandleFunc("/logout", auth.Logout)
	mux.HandleFunc("/signup", auth.Signup)
	mux.HandleFunc("/categories", post.ServeCategories)
	mux.HandleFunc("/upload-image", auth.Middleware(http.HandlerFunc(post.UploadImage)))

	mux.HandleFunc("/view-post", post.ViewPost)
	mux.HandleFunc("/category", post.ViewPostsByCategory)

	mux.HandleFunc("/userfilter", auth.Middleware(http.HandlerFunc(post.FilterbyUser)))
	mux.HandleFunc("/likesfilter", auth.Middleware(http.HandlerFunc(post.FilterbyLikes)))

	// Comment Routes
	mux.HandleFunc("/comments", comments.GetComments)
	mux.HandleFunc("/comments/create", auth.Middleware(http.HandlerFunc(comments.CreateComment)))
	mux.HandleFunc("/create-post-form", auth.Middleware(http.HandlerFunc(post.ServeCreatePostForm)))
	mux.HandleFunc("/create-post", auth.Middleware(http.HandlerFunc(post.CreatePost)))
	mux.HandleFunc("/post/react", auth.Middleware(http.HandlerFunc(post.ReactToPost)))
	mux.HandleFunc("/comments/react", auth.Middleware(http.HandlerFunc(comments.ReactToComment)))
	// static
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	

	fmt.Println("Server running http://localhost:8080/  and go to /login to login")
	http.ListenAndServe(":8080", mux)
}
