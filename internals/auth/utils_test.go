package auth

import (
	"testing"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"Valid email", "test@example.com", true},
		{"Invalid email - no @", "testexample.com", false},
		{"Invalid email - no domain", "test@", false},
		{"Invalid email - invalid chars", "test!@example.com", false},
		{"Invalid email - too long", string(make([]byte, 255)) + "@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidEmail(tt.email); got != tt.want {
				t.Errorf("isValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"Valid username", "john123", true},
		{"Valid username with underscore", "john_doe", true},
		{"Valid username with hyphen", "john-doe", true},
		{"Invalid - too short", "jo", false},
		{"Invalid - too long", "thisusernameiswaytoolongtobevalid123", false},
		{"Invalid - special chars", "john@doe", false},
		{"Invalid - numbers only", "12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidUsername(tt.username); got != tt.want {
				t.Errorf("isValidUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordEncryptionAndDecryption(t *testing.T) {
	password := "myPassword123"

	// Test encryption
	hashedPassword, err := encryptPassword(password)
	if err != nil {
		t.Errorf("encryptPassword() error = %v", err)
		return
	}
	if hashedPassword == password {
		t.Error("encryptPassword() failed: hashed password equals original password")
	}

	// Test decryption
	if !decryptPassword(hashedPassword, password) {
		t.Error("decryptPassword() failed: could not verify password")
	}

	// Test wrong password
	if decryptPassword(hashedPassword, "wrongPassword") {
		t.Error("decryptPassword() failed: verified wrong password")
	}
}
