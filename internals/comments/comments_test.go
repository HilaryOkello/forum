package comments

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"forum/db"
	"forum/internals/auth"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetPostComments(t *testing.T) {
	// Setup test database
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create required tables
	setupSQL := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL
		);

	CREATE TABLE comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,     
		post_id INTEGER NOT NULL,                 
		user_id INTEGER NOT NULL, 
		parent_id INTEGER DEFAULT NULL,                
		content TEXT NOT NULL,                    
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES posts(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
		FOREIGN KEY (parent_id) REFERENCES comments(id)
	);

		CREATE TABLE comment_reactions (
			comment_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			reaction_type TEXT NOT NULL,
			FOREIGN KEY (comment_id) REFERENCES comments(id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			PRIMARY KEY (comment_id, user_id)
		);
	`
	_, err = testDB.Exec(setupSQL)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test data
	testData := `
		INSERT INTO users (id, username) VALUES 
			(1, 'user1'),
			(2, 'user2'),
			(3, 'user3');

		INSERT INTO comments (id, post_id, user_id, parent_id, content, created_at) VALUES
			(1, '123', 1, NULL, 'Root comment 1', '2024-01-28 10:00:00'),
			(2, '123', 2, 1, 'Reply to root 1', '2024-01-28 10:01:00'),
			(3, '123', 3, NULL, 'Root comment 2', '2024-01-28 10:02:00'),
			(4, '123', 1, 2, 'Reply to reply', '2024-01-28 10:03:00');

		INSERT INTO comment_reactions (comment_id, user_id, reaction_type) VALUES
			(1, 2, 'LIKE'),
			(1, 3, 'LIKE'),
			(2, 1, 'DISLIKE'),
			(3, 2, 'LIKE');
	`
	_, err = testDB.Exec(testData)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Temporarily replace the global DB with our test DB
	originalDB := db.DB
	db.DB = testDB
	defer func() { db.DB = originalDB }()

	// Test cases
	tests := []struct {
		name          string
		postID        string
		userID        int64
		expectedCount int
		checkFunc     func(*testing.T, []Comment) bool
	}{
		{
			name:          "Valid post with comments",
			postID:        "123",
			userID:        1,
			expectedCount: 2, // Expected root comments
			checkFunc: func(t *testing.T, comments []Comment) bool {
				if len(comments) != 2 {
					t.Errorf("Expected 2 root comments, got %d", len(comments))
					return false
				}

				// Check first root comment
				if comments[0].ID != 3 ||
					comments[0].Content != "Root comment 2" ||
					comments[0].Likes != 1 ||
					comments[0].UserReaction != nil {
					t.Error("Root comment 2 data mismatch")
					return false
				}

				if len(comments[0].Children) != 0 {
					t.Errorf("Expected 0 child for Root comment 2, got %d", len(comments[0].Children))
					return false
				}

				// Check nested reply
				firstReply := comments[1].Children[0]
				dislikeReaction := "DISLIKE"
				if firstReply.ID != 2 ||
					firstReply.Content != "Reply to root 1" ||
					!reflect.DeepEqual(firstReply.UserReaction, &dislikeReaction) {
					t.Error("First reply data mismatch")
					return false
				}

				// Check second root comment
				if comments[1].ID != 1 ||
					comments[1].Content != "Root comment 1" ||
					comments[1].Likes != 2 ||
					len(comments[1].Children) == 0 {
					t.Error("Second root comment data mismatch", comments[1])
					return false
				}

				return true
			},
		},
		{
			name:          "Empty post",
			postID:        "nonexistent",
			userID:        1,
			expectedCount: 0,
			checkFunc: func(t *testing.T, comments []Comment) bool {
				if len(comments) != 0 {
					t.Errorf("Expected no comments, got %d", len(comments))
					return false
				}
				return true
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comments, err := getPostComments(tt.postID, tt.userID)
			if err != nil {
				t.Fatalf("getPostComments returned unexpected error: %v", err)
			}
			if len(comments) != tt.expectedCount {
				t.Errorf("Expected %d comments, got %d", tt.expectedCount, len(comments))
			}
			if tt.checkFunc != nil {
				if !tt.checkFunc(t, comments) {
					t.Error("Check function failed")
				}
			}
		})
	}
}

func TestGetComments(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		postID        string
		userID        int64
		loggedIn      bool
		expectedCode  int
		expectedCount int
	}{
		{
			name:          "Get comments for post - logged in user",
			method:        http.MethodGet,
			postID:        "1",
			userID:        1,
			loggedIn:      true,
			expectedCode:  http.StatusOK,
			expectedCount: 2,
		},
		{
			name:          "Get comments for post - not logged in",
			method:        http.MethodGet,
			postID:        "1",
			loggedIn:      false,
			expectedCode:  http.StatusOK,
			expectedCount: 2,
		},
		{
			name:         "Invalid method",
			method:       http.MethodPost,
			postID:       "1",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "Missing post_id",
			method:       http.MethodGet,
			postID:       "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:          "Post with no comments",
			method:        http.MethodGet,
			postID:        "999",
			expectedCode:  http.StatusOK,
			expectedCount: 0,
		},
	}
	setupTestDB := func(t *testing.T) *testDB {
		db, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			t.Fatalf("Failed to open test database: %v", err)
		}

		// Create necessary tables
		_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL
		);
	
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,     
			post_id INTEGER NOT NULL,                 
			user_id INTEGER NOT NULL, 
			parent_id INTEGER DEFAULT NULL,                
			content TEXT NOT NULL,                    
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
			FOREIGN KEY (parent_id) REFERENCES comments(id)
		);
		CREATE TABLE comment_reactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,     
			comment_id INTEGER NOT NULL,              
			user_id INTEGER NOT NULL,                 
			reaction_type TEXT CHECK (reaction_type IN ('LIKE', 'DISLIKE')) NOT NULL,
			FOREIGN KEY (comment_id) REFERENCES comments(id),
			FOREIGN KEY (user_id) REFERENCES users(id) 
			UNIQUE (comment_id, user_id)
		);
		`)
		if err != nil {
			t.Fatalf("Failed to create test tables: %v", err)
		}

		return &testDB{db}
	}

	seedTestData := func(t *testing.T, db *testDB) {
		// Insert test users
		_, err := db.Exec(`
			INSERT INTO users (id, username) VALUES 
				(1, 'testuser1'),
				(2, 'testuser2'),
				(3, 'testuser3');
		`)
		if err != nil {
			t.Fatalf("Failed to seed users: %v", err)
		}

		// Insert test comments
		_, err = db.Exec(`
			INSERT INTO comments (id, post_id, user_id, parent_id, content, created_at)
			VALUES 
				(1, 1, 1, NULL, 'First comment', CURRENT_TIMESTAMP),
				(2, 1, 2, NULL, 'Second comment', CURRENT_TIMESTAMP),
				(3, 1, 3, 1, 'Reply to first comment', CURRENT_TIMESTAMP),
				(4, 2, 1, NULL, 'Comment on different post', CURRENT_TIMESTAMP);
		`)
		if err != nil {
			t.Fatalf("Failed to seed comments: %v", err)
		}

		// Insert test reactions
		_, err = db.Exec(`
			INSERT INTO comment_reactions (comment_id, user_id, reaction_type)
			VALUES
				(1, 2, 'LIKE'),
				(1, 3, 'LIKE'),
				(2, 1, 'DISLIKE'),
				(3, 1, 'LIKE');
		`)
		if err != nil {
			t.Fatalf("Failed to seed reactions: %v", err)
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			testDB := setupTestDB(t)
			defer testDB.Close()

			// Replace the global db with our test db
			db.DB = testDB.DB

			// Seed test data
			seedTestData(t, testDB)

			// Create request
			url := "/comments"
			if tt.postID != "" {
				url += "?post_id=" + tt.postID
			}
			req := httptest.NewRequest(tt.method, url, nil)
			rec := httptest.NewRecorder()

			// Call the handler directly since we're bypassing auth middleware
			GetComments(rec, req)

			// Check status code
			if rec.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, rec.Code)
			}

			// For successful requests, verify the response
			if tt.expectedCode == http.StatusOK {
				var comments []Comment // You'll need to define this struct based on your implementation
				t.Log(rec.Body.String())
				err := json.NewDecoder(rec.Body).Decode(&comments)
				if err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if len(comments) != tt.expectedCount {
					t.Errorf("Expected %d comments, got %d", tt.expectedCount, len(comments))
				}

				// Test specific comment attributes
				if len(comments) > 0 {
					// Verify first comment has correct post_id
					testPostID, _ := strconv.Atoi(tt.postID)
					if tt.postID != "" && comments[0].PostID != int64(testPostID) {
						t.Errorf("Expected post_id %s, got %v", tt.postID, comments[0].PostID)
					}

					// If user is logged in, verify reaction data is present
					if tt.loggedIn {
						// Add specific checks for reaction data based on your Comment struct
						// For example:
						// if comments[0].UserReaction == nil {
						//     t.Error("Expected user reaction data to be present")
						// }
					}
				}
			}
		})
	}
}

type testDB struct {
	*sql.DB
}

type MockSession struct {
	UserID int64
}

func setupTestDB(t *testing.T) *testDB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the necessary tables
	_, err = db.Exec(`
	CREATE TABLE comment_reactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,     
		comment_id INTEGER NOT NULL,              
		user_id INTEGER NOT NULL,                 
		reaction_type TEXT CHECK (reaction_type IN ('LIKE', 'DISLIKE')) NOT NULL,
		FOREIGN KEY (comment_id) REFERENCES comments(id),
		FOREIGN KEY (user_id) REFERENCES users(id) 
		UNIQUE (comment_id, user_id)
	);
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return &testDB{db}
}

func TestReactToComment(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		userID           int64
		input            reactToCommentInput
		existingReaction string
		expectedStatus   int
		expectedResponse map[string]string
	}{
		{
			name:   "Add new reaction",
			method: http.MethodPost,
			userID: 1,
			input: reactToCommentInput{
				CommentID:    1,
				ReactionType: "LIKE",
			},
			existingReaction: "",
			expectedResponse: map[string]string{
				"status":           "added",
				"updatedReaction":  "LIKE",
				"previousReaction": "",
			},
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			userID:         1,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Invalid input - missing commentID",
			method: http.MethodPost,
			userID: 1,
			input: reactToCommentInput{
				ReactionType: "LIKE",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid input - missing reactionType",
			method: http.MethodPost,
			userID: 1,
			input: reactToCommentInput{
				CommentID: 1,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			testDB := setupTestDB(t)
			defer testDB.Close()

			// Replace the global db with our test db
			db.DB = testDB.DB

			// Setup existing reaction if needed
			if tt.existingReaction != "" {
				_, err := testDB.Exec(
					"INSERT INTO comment_reactions (comment_id, user_id, reaction_type) VALUES (?, ?, ?)",
					tt.input.CommentID,
					tt.userID,
					tt.existingReaction,
				)
				if err != nil {
					t.Fatalf("Failed to setup test data: %v", err)
				}
			}

			// Create request
			var body bytes.Buffer
			if tt.method == http.MethodPost {
				err := json.NewEncoder(&body).Encode(tt.input)
				if err != nil {
					t.Fatalf("Failed to encode request body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/react", &body)
			rec := httptest.NewRecorder()

			// Add session to context
			session := &MockSession{UserID: tt.userID}
			ctx := context.WithValue(req.Context(), auth.UserSessionKey, session)
			req = req.WithContext(ctx)

			// Call the handler
			ReactToComment(rec, req)

			// For successful requests, verify the response
			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				t.Error(rec.Body.String())
				err := json.NewDecoder(rec.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				for key, expected := range tt.expectedResponse {
					if got := response[key]; got != expected {
						t.Errorf("Expected %s to be %q, got %q", key, expected, got)
					}
				}

				// Verify database state
				var count int
				err = testDB.QueryRow("SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND user_id = ?",
					tt.input.CommentID, tt.userID).Scan(&count)
				if err != nil {
					t.Fatalf("Failed to query database: %v", err)
				}

				expectedCount := 1
				if tt.expectedResponse["status"] == "removed" {
					expectedCount = 0
				}

				if count != expectedCount {
					t.Errorf("Expected %d reactions in database, got %d", expectedCount, count)
				}
			}
		})
	}
}
