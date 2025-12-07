package main

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/jorwong/go_user_accounts/api"
	"github.com/jorwong/go_user_accounts/models"
	"github.com/jorwong/go_user_accounts/pb"
	pkg "github.com/jorwong/go_user_accounts/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	models.InitDB()
	pkg.StartLoggerWorker()
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		selector.UnaryServerInterceptor(
			auth.UnaryServerInterceptor(api.Authenticator),
			selector.MatchFunc(api.AuthMatcher),
		),
		selector.UnaryServerInterceptor(
			api.RateLimter,
			selector.MatchFunc(api.RateLimiterMatcher),
		),
	))
	pb.RegisterUserAccountsServer(s, &api.Server{})
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	//router := mux.NewRouter()
	//
	//router.HandleFunc("/register", api.Register).Methods("POST")
	//router.HandleFunc("/login", api.Login).Methods("POST")
	//router.HandleFunc("/logout", api.Logout).Methods("POST")
	//
	//authenticationSubrouter := router.PathPrefix("/auth").Subrouter()
	//authenticationSubrouter.Use(jwt.VerifyJWT)
	//authenticationSubrouter.HandleFunc("/profile", api.GetProfile).Methods("POST")
	//
	//http.ListenAndServe(":8080", router)
}
