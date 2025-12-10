package interceptors

import (
	"context"
	"fmt"
	"runtime/debug"

	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryInterceptor(logger pkgLogger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				logger.Error(ctx, fmt.Errorf("panic recovered: %v", r), "Panic in gRPC handler", pkgLogger.Tags{
					"method": info.FullMethod,
					"panic":  fmt.Sprintf("%v", r),
					"stack":  string(stack),
				})
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
