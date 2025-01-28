package post

import (
	"database/sql"
	"testing"
	"time"

	"forum/db"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// Mock DB setup helper
func setupMockDB(*testing.T) (sqlmock.Sqlmock, error) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	// Replace the global DB with our mock
	db.DB = mockDB

	return mock, nil
}

func TestFetchPosts(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedPosts []Post
		expectedError string
	}{
		{
			name:   "Success - Multiple posts with reactions",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "title", "content", "username", "created_at",
					"comment_count", "likes", "dislikes", "user_reaction",
				}).
					AddRow(1, "Test Post 1", "Content 1", "user1", time.Now(), 5, 10, 2, "LIKE").
					AddRow(2, "Test Post 2", "Content 2", "user2", time.Now(), 0, 0, 0, "")
				mock.ExpectQuery("^SELECT").WithArgs(1).WillReturnRows(rows)
			},
			expectedPosts: []Post{
				{ID: 1, Title: "Test Post 1", Content: "Content 1", UserName: "user1", CommentCount: 5, Likes: 10, Dislikes: 2, UserReaction: "LIKE"},
				{ID: 2, Title: "Test Post 2", Content: "Content 2", UserName: "user2", CommentCount: 0, Likes: 0, Dislikes: 0, UserReaction: ""},
			},
		},
		{
			name:   "Empty result set",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "title", "content", "username", "created_at",
					"comment_count", "likes", "dislikes", "user_reaction",
				})
				mock.ExpectQuery("^SELECT").WithArgs(1).WillReturnRows(rows)
			},
			expectedPosts: []Post{},
		},
		{
			name:   "Database error",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT").WithArgs(1).WillReturnError(sql.ErrConnDone)
			},
			expectedError: "failed to fetch posts",
		},
		{
			name:   "Scan error - invalid data type",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "title", "content", "username", "created_at",
					"comment_count", "likes", "dislikes", "user_reaction",
				}).AddRow("invalid", "Test", "Content", "user1", time.Now(), 0, 0, 0, "")
				mock.ExpectQuery("^SELECT").WithArgs(1).WillReturnRows(rows)
			},
			expectedError: "failed to scan post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := setupMockDB(t)
			if err != nil {
				t.Fatalf("Failed to setup mock DB: %v", err)
			}
			defer db.DB.Close()

			tt.mockSetup(mock)

			posts, err := FetchPosts(tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedPosts), len(posts))
				for i, expectedPost := range tt.expectedPosts {
					assert.Equal(t, expectedPost.ID, posts[i].ID)
					assert.Equal(t, expectedPost.Title, posts[i].Title)
					assert.Equal(t, expectedPost.UserReaction, posts[i].UserReaction)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFetchPostFromDB(t *testing.T) {
	tests := []struct {
		name          string
		postID        string
		userID        int64
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedPost  *Post
		expectedError string
	}{
		{
			name:   "Success - Post exists with reactions",
			postID: "1",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "title", "content", "username", "created_at",
					"comment_count", "likes", "dislikes", "user_reaction",
				}).AddRow(1, "Test Post", "Content", "user1", time.Now(), 5, 10, 2, "LIKE")
				mock.ExpectQuery("^SELECT").WithArgs(1, "1").WillReturnRows(rows)
			},
			expectedPost: &Post{
				ID: 1, Title: "Test Post", Content: "Content",
				UserName: "user1", CommentCount: 5, Likes: 10,
				Dislikes: 2, UserReaction: "LIKE",
			},
		},
		{
			name:   "Post not found",
			postID: "999",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT").WithArgs(1, "999").WillReturnError(sql.ErrNoRows)
			},
			expectedError: "not found",
		},
		{
			name:   "Invalid post ID",
			postID: "invalid",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT").WithArgs(1, "invalid").WillReturnError(sql.ErrConnDone)
			},
			expectedError: "failed to fetch post",
		},
		{
			name:   "Database connection error",
			postID: "1",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT").WithArgs(1, "1").WillReturnError(sql.ErrConnDone)
			},
			expectedError: "failed to fetch post",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := setupMockDB(t)
			if err != nil {
				t.Fatalf("Failed to setup mock DB: %v", err)
			}
			defer db.DB.Close()

			tt.mockSetup(mock)

			post, err := fetchPostFromDB(tt.postID, tt.userID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, post)
				assert.Equal(t, tt.expectedPost.ID, post.ID)
				assert.Equal(t, tt.expectedPost.Title, post.Title)
				assert.Equal(t, tt.expectedPost.UserReaction, post.UserReaction)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
