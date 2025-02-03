package post

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"forum/internals/auth"
	"forum/internals/fails"
)

var (
	currentUpload = make(map[int64]*ImageUploadResult)
	uploadMutex   sync.Mutex
)

func UploadImage(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		fails.ErrorPageHandler(w, r, http.StatusMethodNotAllowed)
		return
	}
	session := auth.CheckIfLoggedIn(w, r)
	if session == nil {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}
	// Set maximum upload size - 20MB
	const maxUploadSize = 5 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large. Maximum size is 5MB", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate unique filename
	filename, err := generateUniqueFilename(header.Filename)

	if err != nil {
		http.Error(w, "Error processing upload", http.StatusInternalServerError)
		return
	}

	// Ensure upload directory exists
	uploadDir := "static/images"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		http.Error(w, "Error processing upload", http.StatusInternalServerError)
		return
	}

	// Create the file
	filepath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Store the result for this user
	uploadMutex.Lock()
	currentUpload[int64(session.UserID)] = &ImageUploadResult{
		Filename: "/static/images/" + filename,
	}
	fmt.Println("upload: ", currentUpload)
	uploadMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message":"upload successful","filename":"%s"}`, "/static/images/"+filename)
}

func generateUniqueFilename(originalFilename string) (string, error) {
	// Generate 16 random bytes
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Get the file extension from original filename
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if ext == "" {
		ext = ".jpg" // Default extension if none provided
	}

	return hex.EncodeToString(bytes) + ext, nil
}
