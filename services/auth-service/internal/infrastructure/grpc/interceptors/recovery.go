package interceptors

import (
	"context"
	"fmt"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
)

func RecoveryInterceptor(logger pkgLogger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				logger.Error(ctx, pkgErrors.NewInternalServerError("panic recovered in gRPC handler"), "Panic in gRPC handler", pkgLogger.Tags{
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
