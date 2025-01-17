package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"forum/db"
	"forum/internals/auth"
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
	mux.HandleFunc("/", auth.ServeHomePage)

	mux.HandleFunc("/login", auth.Login)
	mux.HandleFunc("/logout", auth.Logout)
	mux.HandleFunc("/signup", auth.Signup)
	mux.HandleFunc("/create-post-form", post.ServeCreatePostForm)
	mux.HandleFunc("/categories", post.ServeCategories)
	mux.HandleFunc("/create-post", post.CreatePost)

	// static
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// // private routes
	// protected := http.NewServeMux()
	// protected.HandleFunc("/dashboard", auth.Dashboard)

	// mux.Handle("/dashboard", auth.Middleware(protected))

	// openBrowser("http://localhost:8080/")
	fmt.Println("Server running http://localhost:8080/  and go to /login to login")
	http.ListenAndServe(":8080", mux)
}

func openBrowser(url string) {
	var err error
	fmt.Println(runtime.GOOS)
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}
}
