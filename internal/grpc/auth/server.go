package authgrpc

import (
	"context"
	"errors"
	"regexp"
	"vieo/auth/internal/lib/jwt"
	"vieo/auth/internal/services/auth"
	"vieo/auth/internal/storage"

	desc "github.com/Avalance-rl/contract-vieo/pkg/auth_v1"
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
		accessToken string,
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

func (s *serverAPI) Login(
	ctx context.Context,
	req *desc.LoginRequest,
) (*desc.LoginResponse, error) {
	if !isEmailValid(req.Email) || !isPasswordValid(req.GetPassword()) || req.GetDeviceAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "not valid email or password")
	}
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetDeviceAddress())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, storage.ErrDeviceLimitExceeded) {
			return nil, status.Error(codes.ResourceExhausted, "device limit exceeded")
		}
		if errors.Is(err, auth.ErrWrongPassword) {
			return nil, status.Error(codes.Unauthenticated, "wrong password")
		}
		// in theory this case should not happen, because either the interceptor will intercept the user without an access token and
		// throw it to the refresh line or front
		if errors.Is(err, storage.ErrDeviceAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "device already exists")
		}
		return nil, status.Error(codes.Internal, "internal server error")

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
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &desc.RegisterResponse{UserId: uid}, nil
}

func (s *serverAPI) RefreshToken(
	ctx context.Context,
	req *desc.RefreshTokenRequest,
) (*desc.RefreshTokenResponse, error) {
	if req.DeviceAddress == "=" || req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "device addres or access token is empty")
	}

	token, err := s.auth.RefreshToken(ctx, req.GetDeviceAddress(), req.GetAccessToken())
	if err != nil {
		if errors.Is(err, jwt.ErrInvalidToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		if errors.Is(err, jwt.ErrIncorrectExpiration) {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		if errors.Is(err, jwt.ErrInvalidSignMethod) {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		if errors.Is(err, jwt.ErrFailedToExtractData) {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		if errors.Is(err, jwt.ErrIncorrectExpiration) {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		if errors.Is(err, storage.ErrDeviceNotFound) {
			return nil, status.Error(codes.NotFound, "device not found")
		}
		if errors.Is(err, auth.ErrAddressMismatch) {
			return nil, status.Error(codes.FailedPrecondition, "address mismatch")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &desc.RefreshTokenResponse{Token: token}, nil
}
func (s *serverAPI) CheckToken(
	context.Context,
	*desc.CheckTokenRequest,
) (*desc.CheckTokenResponse, error) {

	return &desc.CheckTokenResponse{Message: "OK"}, nil
}

func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}

func isPasswordValid(p string) bool {
	passwordRegex := regexp.MustCompile(`^[a-z0-9]{8,14}$`)
	return passwordRegex.MatchString(p)
}
