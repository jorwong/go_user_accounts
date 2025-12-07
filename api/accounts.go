package api

import (
	"context"
	"github.com/jorwong/go_user_accounts/models"
	"github.com/jorwong/go_user_accounts/pb"
	jwt "github.com/jorwong/go_user_accounts/pkg/jwt"
	pkg "github.com/jorwong/go_user_accounts/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type Server struct {
	pb.UserAccountsServer
}

func (s *Server) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterReply, error) {

	if in.Name == "" || in.Email == "" || in.Password == "" {
		// If any required field is empty
		return nil, status.Errorf(codes.InvalidArgument, "Missing Argument")
	}

	err := models.CreateUser(in.Email, in.Name, in.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal Server Error")
	}

	return &pb.RegisterReply{
		Message: "User Registered Successfully",
	}, nil
}

func (s *Server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginReply, error) {

	if in.Email == "" || in.Password == "" {
		pkg.LogChannel <- time.Now().String() + "," + "Missing required fields (Email, or Password)."
		return nil, status.Errorf(codes.InvalidArgument, "Missing required fields (Email, or Password).")
	}

	foundUser, err := models.FindUserByEmail(in.Email)

	if err != nil && err.Error() == "DB_ERROR" {
		pkg.LogChannel <- time.Now().String() + "," + "DB ERROR"
		return nil, status.Errorf(codes.Internal, "Internal Server Error.")
	}

	if foundUser == nil || !foundUser.CheckPasswordHash(in.Password) {
		pkg.LogChannel <- time.Now().String() + "," + "Invalid Credentials for " + in.Email

		return nil, status.Errorf(codes.Unauthenticated, "Invalid Credentials")
	}

	//ifIsAllowed, err := ratelimit.IsAllowed(foundUser.Email)
	//if err != nil && err.Error() != "RATE_LIMITED" {
	//	pkg.LogChannel <- time.Now().String() + "," + err.Error()
	//	return nil, status.Errorf(codes.Internal, "ERROR :"+err.Error())
	//}
	//
	//if !ifIsAllowed {
	//	pkg.LogChannel <- time.Now().String() + "," + "Rate Limited for " + in.Email
	//
	//	return nil, status.Errorf(codes.ResourceExhausted, "Rate Limited")
	//}

	jwtToken, err := jwt.GenerateJWT(foundUser.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ERROR :"+err.Error())
	}

	pkg.LogChannel <- time.Now().String() + "," + "Successful Login for " + foundUser.Email
	return &pb.LoginReply{Message: "TOKEN: " + jwtToken}, nil
}

func (s *Server) Logout(ctx context.Context, in *pb.LogOutRequest) (*pb.LogOutReply, error) {

	if in.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Missing required fields (Email).")
	}

	foundUser, err := models.FindUserByEmail(in.Email)

	if err != nil && err.Error() == "DB_ERROR" {
		return nil, status.Errorf(codes.Internal, "Internal Server Error.")
	}

	err = models.RevokeSession(foundUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal Server Error.")
	}

	return &pb.LogOutReply{}, nil
}

func (s *Server) Profile(ctx context.Context, in *pb.ProfileRequest) (*pb.ProfileReply, error) {

	if in.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Missing required fields (Email).")
	}

	User, err := models.FindUserByEmail(in.Email)
	if err != nil || User == nil {
		return nil, status.Errorf(codes.Internal, "Internal Server Error.")
	}

	return &pb.ProfileReply{Message: User.Email + "|" + User.Name}, nil
}
