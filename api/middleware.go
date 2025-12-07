package api

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/jorwong/go_user_accounts/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

const (
	INVALID_SESSION = "INVALID_SESSION"
	ERROR           = "ERROR"
	VALID           = "VALID"
)

var secret = []byte("secret")

func CheckIfUserSessionIsValid(session string, email string) (string, bool) {
	sessionKey, err := models.GetTokenFromRedis(email)
	if err != nil {
		return ERROR, false
	}

	if sessionKey != session {
		return INVALID_SESSION, false
	}

	return VALID, true
}

func verifyJWT(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return the secret key for verification
		return secret, nil // <-- FIX: Use the defined secret
	})

	if err != nil {
		fmt.Printf("JWT validation error: %v\n", err)
		return false, err
	}

	if !token.Valid {
		return false, nil
	}

	return true, nil
}

func Authenticator(ctx context.Context) (context.Context, error) {
	tokenString, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	result, err := verifyJWT(tokenString)

	if !result || err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalide Auth Token")
	}

	return ctx, nil
}

// /PackageName.ServiceName/MethodName
var authenticatedMethods = map[string]bool{
	"/api.UserAccounts/Logout":  true,
	"/api.UserAccounts/Profile": true,
}

func AuthMatcher(ctx context.Context, callMeta interceptors.CallMeta) bool { // Check if the current full method name exists in the map
	_, requiresAuth := authenticatedMethods[callMeta.FullMethod()]
	fmt.Println(callMeta.FullMethod())
	fmt.Println("requiresAuth:" + strconv.FormatBool(requiresAuth))
	// Return true if the method is found in the map
	return requiresAuth
}
