package interceptors

import (
	"context"
	"time"

	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LoggingInterceptor(logger pkgLogger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		md, _ := metadata.FromIncomingContext(ctx)
		requestID := extractRequestID(md)

		logger.Info(ctx, "gRPC request started", pkgLogger.Tags{
			"method":     info.FullMethod,
			"request_id": requestID,
		})

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			logger.Error(ctx, err, "gRPC request failed", pkgLogger.Tags{
				"method":      info.FullMethod,
				"duration_ms": duration.Milliseconds(),
				"request_id":  requestID,
			})
		} else {
			logger.Info(ctx, "gRPC request completed", pkgLogger.Tags{
				"method":      info.FullMethod,
				"duration_ms": duration.Milliseconds(),
				"request_id":  requestID,
			})
		}

		return resp, err
	}
}

func extractRequestID(md metadata.MD) string {
	if vals := md.Get("x-request-id"); len(vals) > 0 {
		return vals[0]
	}
	return "unknown"
}
