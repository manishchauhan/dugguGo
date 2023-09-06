// passwordutils/passwordutils.go
package passwordutils

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a plaintext password and returns the hashed value.
func HashPassword(plaintextPassword string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
}

// ComparePasswords compares a stored hashed password with an input password.
// Returns nil if they match, or an error if they don't.
func ComparePasswords(hashedPassword []byte, inputPassword string) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(inputPassword))
}
