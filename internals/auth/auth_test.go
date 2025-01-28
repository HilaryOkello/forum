package auth

import (
	"database/sql"
	"testing"
	"time"

	"forum/db"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Create an in-memory SQLite database for testing"
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the users table
	_, err = testDB.Exec(`
        CREATE TABLE users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL,
            created_at DATETIME NOT NULL
        )
    `)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Insert test data
	stmt, err := testDB.Prepare(`
        INSERT INTO users (username, email, password, created_at)
        VALUES (?, ?, ?, ?)
    `)
	if err != nil {
		t.Fatalf("Failed to prepare insert statement: %v", err)
	}
	defer stmt.Close()

	// Insert test users
	testUsers := []struct {
		username string
		email    string
		password string
	}{
		{"testuser1", "test1@example.com", "hashedpassword1"},
		{"testuser2", "test2@example.com", "hashedpassword2"},
	}

	createdTime := time.Now().Format("2006-01-02 15:04:05")
	for _, user := range testUsers {
		_, err := stmt.Exec(user.username, user.email, user.password, createdTime)
		if err != nil {
			t.Fatalf("Failed to insert test user: %v", err)
		}
	}

	return testDB
}

func TestReadfromDb(t *testing.T) {
	// Set up test database
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Replace the production database with our test database
	originalDB := db.DB
	db.DB = testDB
	defer func() { db.DB = originalDB }()

	// Call the function to test
	users := ReadfromDb()

	// Check the results
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	// Define expected usernames
	expectedUsers := map[string]string{
		"testuser1": "test1@example.com",
		"testuser2": "test2@example.com",
	}

	// Check each user
	for _, user := range users {
		expectedEmail, exists := expectedUsers[user.UserName]
		if !exists {
			t.Errorf("Unexpected username: %s", user.UserName)
			continue
		}
		if user.Email != expectedEmail {
			t.Errorf("For user %s: expected email %s, got %s",
				user.UserName, expectedEmail, user.Email)
		}
	}
}

func TestSaveUserToDb(t *testing.T) {
	// Set up test database
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Replace the production database with our test database
	originalDB := db.DB
	db.DB = testDB
	defer func() { db.DB = originalDB }()

	// Create a test user
	newUser := User{
		UserName: "newuser",
		Email:    "new@example.com",
		Password: "hashedpassword3",
	}

	// Save the user
	err := SaveUserToDb(newUser)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	// Verify the user was saved
	var savedUser User
	err = testDB.QueryRow(`
        SELECT id, username, email, password
        FROM users
        WHERE email = ?
    `, newUser.Email).Scan(&savedUser.ID, &savedUser.UserName, &savedUser.Email, &savedUser.Password)
	if err != nil {
		t.Fatalf("Failed to retrieve saved user: %v", err)
	}

	// Check the saved data
	if savedUser.UserName != newUser.UserName {
		t.Errorf("Expected username %s, got %s", newUser.UserName, savedUser.UserName)
	}
	if savedUser.Email != newUser.Email {
		t.Errorf("Expected email %s, got %s", newUser.Email, savedUser.Email)
	}

	// Test duplicate email
	err = SaveUserToDb(newUser)
	if err == nil {
		t.Error("Expected error when saving duplicate email, got nil")
	}
}
