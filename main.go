package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"forum/db"
	"forum/internals"
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
	mux.HandleFunc("/", internals.Index)

	mux.HandleFunc("/login", internals.Login)
	mux.HandleFunc("/logout", internals.Logout)
	mux.HandleFunc("/signup", internals.Signup)

	//static 
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// // private routes
	// protected := http.NewServeMux()
	// protected.HandleFunc("/dashboard", internals.Dashboard)

	// mux.Handle("/dashboard", internals.Middleware(protected))

    openBrowser("http://localhost:8080/")
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
