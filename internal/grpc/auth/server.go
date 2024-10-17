package authgrpc

import (
	"context"
	"errors"
	"regexp"

	desc "github.com/Avalance-rl/contract-vieo/pkg/auth_v1"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Auth interface for the service layer
type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		deviceAddress string,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID int64, err error)
	RefreshToken(
		ctx context.Context,
		deviceAddress string,
	) (token string, err error)
}

// serverAPI handles requests
type serverAPI struct {
	desc.UnimplementedAuthServer //
	auth                         Auth
}

// Register processes requests that come to the grpc server
func Register(gRPC *grpc.Server, auth Auth) {
	desc.RegisterAuthServer(gRPC, &serverAPI{auth: auth}) // регистрация обработчика
}

// TODO: дописать логику

func (s *serverAPI) Login(
	ctx context.Context,
	req *desc.LoginRequest,
) (*desc.LoginResponse, error) {
	// TODO: сделать валидацию поступающих данных нормальную, есть пакеты специальные
	if !isEmailValid(req.Email) || !isPasswordValid(req.GetPassword()) {
		return nil, status.Error(codes.InvalidArgument, "not valid email or password")
	}
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.DeviceAddress)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}
	return &desc.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *desc.RegisterRequest,
) (*desc.RegisterResponse, error) {
	if !isEmailValid(req.Email) || !isPasswordValid(req.GetPassword()) {
		return nil, status.Error(codes.InvalidArgument, "not valid email or password")
	}
	uid, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, status.Error(codes.AlreadyExists, "user already exists")
			}
		} // unique_violation
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &desc.RegisterResponse{UserId: uid}, nil
}

func (s *serverAPI) RefreshToken(
	ctx context.Context,
	req *desc.RefreshTokenRequest,
) (*desc.RefreshTokenResponse, error) {

	return &desc.RefreshTokenResponse{Token: ""}, nil
}

func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}

func isPasswordValid(p string) bool {
	passwordRegex := regexp.MustCompile(`^[a-z0-9]{8,14}$`)
	return passwordRegex.MatchString(p)
}
