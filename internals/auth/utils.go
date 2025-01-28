package auth

import (
	"regexp"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func isValidEmail(email string) bool {
	// Basic email regex pattern
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, err := regexp.MatchString(pattern, email)
	if err != nil {
		return false
	}
	return match && len(email) <= 254 // Standard max email length
}

func isValidUsername(username string) bool {
	// Check length (for example, between 3 and 30 characters)
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	// Check if contains only allowed characters
	for _, char := range username {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) && char != '_' && char != '-' {
			return false
		}
	}

	containsLetter := false
	for _, char := range username {
		if unicode.IsLetter(char) {
			containsLetter = true
			break
		}
	}

	return containsLetter
}

func encryptPassword(password string) (string, error) {
	bcryptPassword, error := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if error != nil {
		return "", error
	}
	return string(bcryptPassword), nil
}

func decryptPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
