package authgrpc

import (
	"context"
	"time"
	"vieo/auth/internal/lib/jwt"
	"vieo/auth/internal/lib/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	secretKey string
	logger    *logger.Logger
}

func NewAuthInterceptor(
	secretKey string,
	logger *logger.Logger,
) *AuthInterceptor {
	return &AuthInterceptor{
		secretKey: secretKey,
		logger:    logger,
	}
}

func (interceptor *AuthInterceptor) Authorize() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		protectedMethods := map[string]bool{
			// here are the methods for which this interceptor is called
			"/auth_v1.Auth/CheckToken": true,
		}
		if protectedMethods[info.FullMethod] {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
			}

			var accessToken string
			if val, ok := md["authorization"]; ok && len(val) > 0 {
				accessToken = val[0]
			} else {
				return nil, status.Errorf(codes.Unauthenticated, "token is not provided")
			}
			_, _, err := jwt.DecodeToken(interceptor.secretKey, accessToken)
			if err != nil {
				return nil, status.Errorf(codes.PermissionDenied, "token is not valid")
			}
		}

		return handler(ctx, req)
	}
}

func (interceptor *AuthInterceptor) Logger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		const op = "auth.AuthInterceptor.Logger"
		log := interceptor.logger.With(
			zap.String("op", op),
			zap.String("method", info.FullMethod),
		)
		startTime := time.Now()

		resp, err = handler(ctx, req)

		log = log.With(
			zap.Duration("duration", time.Since(startTime)),
		)

		if err != nil {
			log.Error("method failed", zap.Error(err))
		} else {
			log.Info("method completed successfully")
		}

		return resp, err
	}
}
