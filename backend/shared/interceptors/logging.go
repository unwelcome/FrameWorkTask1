package interceptors

import (
	"context"
	"path"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// OperationIDMetaKey is the gRPC metadata key used to propagate the operation
// ID from the gateway to downstream services.
const OperationIDMetaKey = "x-op-id"

// NewLoggingInterceptor returns a gRPC unary server interceptor that logs
// every method call with the correct zerolog level based on the gRPC status code:
//
//   - OK                                          → Info
//   - NotFound, InvalidArgument, PermissionDenied,
//     FailedPrecondition, AlreadyExists,
//     Unauthenticated, Canceled                  → Warn
//   - everything else (Internal, Unavailable …)  → Error
//
// logger should be the httpLogger returned by logger.Setup — a logger without
// the Caller hook, so interceptor log lines don't show the useless
// "interceptors/logging.go:XX" caller. The "caller" field remains present in
// all other application logs (startup, fatal errors, etc.) via the appLogger.
//
// The operation ID is read from incoming gRPC metadata (key: x-operation-id).
func NewLoggingInterceptor(logger zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Milliseconds()

		// Read operation ID from incoming metadata set by the gateway.
		operationID := OperationIDFromContext(ctx)

		// info.FullMethod is "/package.Service/Method" — take only "Method".
		method := path.Base(info.FullMethod)
		code := status.Code(err)

		// time is added as the first explicit field so it appears right after
		// "level" — same pattern used in the HTTP request logger middleware.
		e := logger.WithLevel(grpcCodeToZerologLevel(code)).
			Time("time", time.Now()).
			Str("id", operationID).
			Str("method", method).
			Int64("duration", duration)

		if err != nil {
			e = e.Str("error", status.Convert(err).Message())
			e.Msg(code.String())
		} else {
			e.Msg("OK")
		}

		return resp, err
	}
}

// OperationIDFromContext извлекает operation ID из входящих gRPC-метаданных
func OperationIDFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(OperationIDMetaKey); len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}

func grpcCodeToZerologLevel(code codes.Code) zerolog.Level {
	switch code {
	case codes.OK:
		return zerolog.InfoLevel
	case codes.NotFound,
		codes.InvalidArgument,
		codes.PermissionDenied,
		codes.FailedPrecondition,
		codes.AlreadyExists,
		codes.Unauthenticated,
		codes.Canceled:
		return zerolog.WarnLevel
	default:
		// Internal, Unavailable, DeadlineExceeded, Unknown, etc.
		return zerolog.ErrorLevel
	}
}
