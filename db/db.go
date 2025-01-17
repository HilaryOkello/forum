package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Category represents a category entity
type Category struct {
	Name        string
	Description string
}

// global variable for database connection
var DB *sql.DB

// Initialize initializes the database connection and applies the schema
func Initialize() error {
	var err error
	DB, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Ensure the database is accessible
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Run the schema SQL to create tables
	err = applySchema()
	if err != nil {
		return fmt.Errorf("failed to apply schema: %v", err)
	}

	// Ensure default categories exist
	err = EnsureDefaultCategories()
	if err != nil {
		return fmt.Errorf("failed to ensure default categories: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// applySchema applies the SQL schema from the schema.sql file
func applySchema() error {
	schemaContent, err := os.ReadFile("./db/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}

	// Execute the SQL statements from the schema
	_, err = DB.Exec(string(schemaContent))
	if err != nil {
		return fmt.Errorf("failed to execute schema SQL: %v", err)
	}

	log.Println("Schema applied successfully")
	return nil
}

// Close closes the database connection
func Close() {
	err := DB.Close()
	if err != nil {
		log.Fatalf("Error closing database: %v", err)
	}
}

// EnsureDefaultCategories inserts default categories if they don't exist
func EnsureDefaultCategories() error {
	// List of default categories
	defaultCategories := []Category{
		{Name: "Technology", Description: "Posts related to tech trends, innovations, and updates"},
		{Name: "Health", Description: "Topics covering fitness, wellness, and healthcare advice"},
		{Name: "Education", Description: "Discussions on learning, schools, and education systems"},
		{Name: "Entertainment", Description: "Posts about movies, music, and other entertainment topics"},
		{Name: "Travel", Description: "Sharing travel experiences, tips, and destinations"},
		{Name: "Food", Description: "Recipes, cooking tips, and culinary experiences"},
		{Name: "Business", Description: "Business insights, entrepreneurship, and market trends"},
		{Name: "Sports", Description: "Sports news, updates, and discussions on favorite teams"},
		{Name: "Lifestyle", Description: "Lifestyle tips, fashion, and daily living topics"},
		{Name: "Politics", Description: "Discussions on political events and governance issues"},
	}

	// Prepare the query to check and insert categories
	for _, category := range defaultCategories {
		query := `
		INSERT INTO categories (name, description)
		SELECT ?, ?
		WHERE NOT EXISTS (
			SELECT 1 FROM categories WHERE name = ?
		)`
		_, err := DB.Exec(query, category.Name, category.Description, category.Name)
		if err != nil {
			return fmt.Errorf("failed to ensure default category '%s': %v", category.Name, err)
		}
	}

	return nil
}
