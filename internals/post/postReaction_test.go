package post

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"forum/db"
	"forum/internals/auth"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestReactToPost(t *testing.T) {
	tests := []struct {
		name              string
		method            string
		setupAuth         bool
		input             reactToPost
		mockSetup         func(mock sqlmock.Sqlmock)
		expectedStatus    int
		expectedResponse  map[string]string
		expectedDBQueries bool
	}{
		{
			name:      "Success - Add new reaction",
			method:    "POST",
			setupAuth: true,
			input: reactToPost{
				PostID:       1,
				ReactionType: "LIKE",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Expect query to check current reaction
				mock.ExpectQuery(`SELECT reaction_type FROM post_reactions`).
					WithArgs(int64(1), int64(1)).
					WillReturnError(sql.ErrNoRows)

				// Expect insert query
				mock.ExpectExec(`INSERT INTO post_reactions`).
					WithArgs(int64(1), int64(1), "LIKE", "LIKE").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			expectedResponse: map[string]string{
				"status":           "added",
				"updatedReaction":  "LIKE",
				"previousReaction": "",
			},
			expectedDBQueries: true,
		},
		{
			name:      "Success - Remove existing reaction",
			method:    "POST",
			setupAuth: true,
			input: reactToPost{
				PostID:       1,
				ReactionType: "LIKE",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Expect query to check current reaction
				mock.ExpectQuery(`SELECT reaction_type FROM post_reactions`).
					WithArgs(int64(1), int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"reaction_type"}).AddRow("LIKE"))

				// Expect delete query
				mock.ExpectExec(`DELETE FROM post_reactions`).
					WithArgs(int64(1), int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedStatus: http.StatusOK,
			expectedResponse: map[string]string{
				"status":           "removed",
				"updatedReaction":  "LIKE",
				"previousReaction": "LIKE",
			},
			expectedDBQueries: true,
		},
		{
			name:      "Success - Update existing reaction",
			method:    "POST",
			setupAuth: true,
			input: reactToPost{
				PostID:       1,
				ReactionType: "LIKE",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Expect query to check current reaction
				mock.ExpectQuery(`SELECT reaction_type FROM post_reactions`).
					WithArgs(int64(1), int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"reaction_type"}).AddRow("DISLIKE"))

				// Expect update query
				mock.ExpectExec(`INSERT INTO post_reactions`).
					WithArgs(int64(1), int64(1), "LIKE", "LIKE").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			expectedResponse: map[string]string{
				"status":           "updated",
				"updatedReaction":  "LIKE",
				"previousReaction": "DISLIKE",
			},
			expectedDBQueries: true,
		},
		{
			name:              "Invalid Method",
			method:            "GET",
			setupAuth:         true,
			expectedStatus:    http.StatusMethodNotAllowed,
			expectedDBQueries: false,
		},
		{
			name:              "Missing Auth Session",
			method:            "POST",
			setupAuth:         false,
			expectedStatus:    http.StatusUnauthorized,
			expectedDBQueries: false,
		},
		{
			name:              "Invalid JSON Input",
			method:            "POST",
			setupAuth:         true,
			input:             reactToPost{}, // Will be overridden with invalid JSON
			expectedStatus:    http.StatusBadRequest,
			expectedDBQueries: false,
		},
		{
			name:      "Invalid Post ID",
			method:    "POST",
			setupAuth: true,
			input: reactToPost{
				PostID:       0,
				ReactionType: "LIKE",
			},
			expectedStatus:    http.StatusBadRequest,
			expectedDBQueries: false,
		},
		{
			name:      "Database Error - Select Query",
			method:    "POST",
			setupAuth: true,
			input: reactToPost{
				PostID:       1,
				ReactionType: "LIKE",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT reaction_type FROM post_reactions`).
					WithArgs(int64(1), int64(1)).
					WillReturnError(sql.ErrConnDone)
			},
			expectedStatus:    http.StatusInternalServerError,
			expectedDBQueries: true,
		},
		{
			name:      "Database Error - Insert Query",
			method:    "POST",
			setupAuth: true,
			input: reactToPost{
				PostID:       1,
				ReactionType: "LIKE",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT reaction_type FROM post_reactions`).
					WithArgs(int64(1), int64(1)).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectExec(`INSERT INTO post_reactions`).
					WithArgs(int64(1), int64(1), "LIKE", "LIKE").
					WillReturnError(sql.ErrConnDone)
			},
			expectedStatus:    http.StatusInternalServerError,
			expectedDBQueries: true,
		},
		{
			name:      "Database Error - Delete Query",
			method:    "POST",
			setupAuth: true,
			input: reactToPost{
				PostID:       1,
				ReactionType: "LIKE",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT reaction_type FROM post_reactions`).
					WithArgs(int64(1), int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"reaction_type"}).AddRow("LIKE"))

				mock.ExpectExec(`DELETE FROM post_reactions`).
					WithArgs(int64(1), int64(1)).
					WillReturnError(sql.ErrConnDone)
			},
			expectedStatus:    http.StatusInternalServerError,
			expectedDBQueries: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock DB if needed
			var mock sqlmock.Sqlmock
			if tt.expectedDBQueries {
				var err error
				mock, err = setupMockDB(t)
				if err != nil {
					t.Fatalf("Failed to setup mock DB: %v", err)
				}
				defer db.DB.Close()

				if tt.mockSetup != nil {
					tt.mockSetup(mock)
				}
			}

			// Create request
			var body []byte
			var err error
			if tt.method == http.MethodPost {
				if tt.name == "Invalid JSON Input" {
					body = []byte(`{invalid json}`)
				} else {
					body, err = json.Marshal(tt.input)
					if err != nil {
						t.Fatalf("Failed to marshal input: %v", err)
					}
				}
			}
			req := httptest.NewRequest(tt.method, "/react", bytes.NewBuffer(body))

			// Setup auth context if needed
			if tt.setupAuth {
				ctx := context.WithValue(req.Context(), auth.UserSessionKey, &auth.Session{
					UserID:   1,
					UserName: "testuser",
				})
				req = req.WithContext(ctx)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler
			ReactToPost(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// For successful responses, check the JSON response
			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse, response)
			}

			// Verify all DB expectations were met
			if tt.expectedDBQueries && mock != nil {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
