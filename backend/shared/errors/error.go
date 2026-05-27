package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Codes 0..16
//
//	 0 - нет ошибок
//	 1 - CANCELLED
//	 2 - UNKNOWN
//	 3 - INVALID_ARGUMENT
//	 4 - DEADLINE_EXCEEDED
//	 5 - NOT_FOUND
//	 6 - ALREADY_EXISTS
//	 7 - PERMISSION_DENIED
//	 8 - RESOURCE_EXHAUSTED
//	 9 - FAILED_PRECONDITION
//	10 - ABORTED
//	11 - OUT_OF_RANGE
//	12 - UNIMPLEMENTED
//	13 - INTERNAL
//	14 - UNAVAILABLE
//	15 - DATA_LOSS
//	16 - UNAUTHENTICATED

type CodeError struct {
	Code int
	Err  error
	// Msg — публичное сообщение для пользователя.
	// Если пусто, возвращается "internal error".
	Msg string
}

// Public создаёт CodeError с публичным сообщением, видимым пользователю.
func Public(code codes.Code, msg string) CodeError {
	return CodeError{Code: int(code), Err: fmt.Errorf(msg), Msg: msg}
}

// Internal создаёт CodeError для неожиданных ошибок — детали скрыты, пользователь видит "internal error".
func Internal(err error) CodeError {
	return CodeError{Code: int(codes.Internal), Err: err}
}

func (e *CodeError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "empty error"
}

// GRPCError converts the CodeError into a gRPC status error.
// Returns nil when Code == 0 (no error).
func (e CodeError) GRPCError() error {
	if e.Code == 0 {
		return nil
	}

	msg := e.Msg
	if msg == "" {
		msg = "internal error"
	}

	if e.Code < 1 || e.Code > 16 {
		return status.Errorf(codes.Internal, "internal error")
	}

	return status.Errorf(codes.Code(e.Code), msg)
}

// HandleError converts a CodeError into a gRPC status error.
// Deprecated: prefer calling .GRPCError() directly on the CodeError value.
// opID and method are kept for call-site compatibility but are ignored.
func HandleError(errorCode CodeError, opID, method string) error {
	return errorCode.GRPCError()
}
