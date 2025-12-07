package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

var secret = []byte("secret")

// GenerateJWT creates a new JWT token for the given username.
// The token includes the following claims:
//   - "user": the provided username
//   - "authorized": a boolean set to true
//   - "exp": expiration time set to 3 minutes from creation
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

// VerifyJWT is an HTTP middleware that validates JWT tokens from the Authorization header.
// It expects a "Bearer <token>" format and verifies the token signature and expiration.
// If validation fails, it returns an HTTP 401 Unauthorized response.
// If successful, it passes control to the next handler.
func VerifyJWT(endpointHandler http.Handler) http.Handler {

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		// 1. Check for Authorization header
		authorizationHeader := request.Header.Get("Authorization")
		if authorizationHeader == "" {
			http.Error(writer, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// 2. Check for "Bearer <token>" format
		parts := strings.Split(authorizationHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(writer, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// 3. Parse and Validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Return the secret key for verification
			return secret, nil // <-- FIX: Use the defined secret
		})

		// 4. Handle parsing errors (e.g., malformed, expired, invalid signature)
		if err != nil {
			// Log the error for debugging, but return generic unauthorized to the client
			fmt.Printf("JWT validation error: %v\n", err)
			http.Error(writer, "Invalid Token", http.StatusUnauthorized)
			return
		}

		// 5. Check if the token is valid (should be covered by the error check above, but safe)
		if !token.Valid {
			http.Error(writer, "Token is not valid", http.StatusUnauthorized)
			return
		}

		// 6. Token is valid: Pass control to the next handler in the chain
		endpointHandler.ServeHTTP(writer, request)
	})
}
