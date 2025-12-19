package api

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/jorwong/go_user_accounts/pb"
	pkg2 "github.com/jorwong/go_user_accounts/pkg/logging"
	pkg "github.com/jorwong/go_user_accounts/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"time"
)

var secret = []byte("secret")

// /PackageName.ServiceName/MethodName
var authenticatedMethods = map[string]bool{
	"/api.UserAccounts/Logout":  true,
	"/api.UserAccounts/Profile": true,
}

var rateLimitedMethods = map[string]bool{
	"/api.UserAccounts/Profile": true,
	"/api.UserAccounts/Login":   true,
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

func AuthMatcher(ctx context.Context, callMeta interceptors.CallMeta) bool { // Check if the current full method name exists in the map
	_, requiresAuth := authenticatedMethods[callMeta.FullMethod()]
	fmt.Println(callMeta.FullMethod())
	fmt.Println("requiresAuth:" + strconv.FormatBool(requiresAuth))
	// Return true if the method is found in the map
	return requiresAuth
}

func RateLimiterInterceptor(rdb *redis.Client) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		fullMethod := info.FullMethod
		var email string

		// You MUST still use the switch statement to correctly cast the 'req'
		// to the appropriate type for the method and extract the email field.
		switch fullMethod {
		case "/api.UserAccounts/Profile":
			if updateReq, ok := req.(*pb.ProfileRequest); ok {
				// Assuming the type is defined in this package
				email = updateReq.GetEmail()
			}
		case "/api.UserAccounts/Login":
			if postReq, ok := req.(*pb.LoginRequest); ok {
				email = postReq.GetEmail()
			}
		default:
			// This case should ideally not be hit if the matcher is correct,
			// but it's a good safeguard.
			fmt.Printf("Warning: RateLimiterInterceptor triggered for unknown method: %s\n", fullMethod)
		}

		// 3. Perform Rate Limiting Logic using the extracted email
		if email != "" {
			fmt.Printf("Applying rate limit for email: %s\n", email)

			// **ACTUAL RATE LIMITING LOGIC**
			// Example: Check rate limit against Redis
			ifIsAllowed, err := pkg.IsAllowed(email, rdb)
			if err != nil && err.Error() != "RATE_LIMITED" {
				pkg2.LogChannel <- time.Now().String() + "," + err.Error()
				return nil, status.Errorf(codes.Internal, "ERROR :"+err.Error())
			}
			if !ifIsAllowed {
				pkg2.LogChannel <- time.Now().String() + "," + "Rate Limited for " + email

				return nil, status.Errorf(codes.ResourceExhausted, "Rate Limited")
			}
		} else {
			// Handle cases where the email field is missing/empty
			fmt.Printf("Could not extract email for method: %s\n", fullMethod)
		}

		// 4. Proceed to the next handler
		return handler(ctx, req)
	}
}

func RateLimiterMatcher(ctx context.Context, callMeta interceptors.CallMeta) bool {
	// Note: The callMeta type might vary, but for the selector package, it usually has FullMethod().
	_, requiresRateLimit := rateLimitedMethods[callMeta.FullMethod()]
	return requiresRateLimit
}
