package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var secret = []byte("secret")

// GenerateJWT creates a new JWT token for the given username.
// The token includes the following claims:
//   - "user": the provided username
//   - "authorized": a boolean set to true
//   - "exp": expiration time set to 3 minutes from creation
//
// It returns the signed JWT token string, or an error if signing fails.
func GenerateJWT(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(15 * time.Minute).Unix()
	claims["authorized"] = true
	claims["user"] = username

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
