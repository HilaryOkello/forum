package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/db"
	"forum/internals/routes"
)

func main() {
	// Initialize the database
	err := db.Initialize()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	mux := routes.RegisteringRoutes()

	fmt.Println("Server running http://localhost:8080/  and go to /login to login")
	http.ListenAndServe(":8080", mux)
}
