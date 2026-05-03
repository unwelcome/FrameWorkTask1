package errors

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Codes -1..16
// -1 - нет ошибок
//  0 - шаблонная ошибка "internal error"
//  1 - CANCELLED
//  2 - UNKNOWN
//  3 - INVALID_ARGUMENT
//  4 - DEADLINE_EXCEEDED
//  5 - NOT_FOUND
//  6 - ALREADY_EXISTS
//  7 - PERMISSION_DENIED
//  8 - RESOURCE_EXHAUSTED
//  9 - FAILED_PRECONDITION
// 10 - ABORTED
// 11 - OUT_OF_RANGE
// 12 - UNIMPLEMENTED
// 13 - INTERNAL
// 14 - UNAVAILABLE
// 15 - DATA_LOSS
// 16 - UNAUTHENTICATED

type CodeError struct {
	Code int
	Err  error
}

func (e *CodeError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "empty error"
}

// HandleError обрабатывает CodeError, логирует и возвращает готовую gRPC ошибку.
func HandleError(errorCode CodeError, opID, method string) error {

	// Ошибки нет, но код не -1 — ошибка разработчика
	if errorCode.Err == nil && errorCode.Code != -1 {
		log.Error().Str("id", opID).Str("method", method).Err(fmt.Errorf("err is nil but code is not -1")).Msg("error")
	}

	switch errorCode.Code {
	// Нет ошибки
	case -1:
		return nil

	// Шаблонная внутренняя ошибка
	case 0:
		log.Error().Str("id", opID).Str("method", method).Err(errorCode.Err).Msg("error")
		return status.Errorf(codes.Internal, "internal error")

	// Публичные ошибки
	case 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16:
		log.Error().Str("id", opID).Str("method", method).Err(errorCode.Err).Msg("error")
		return status.Errorf(codes.Code(errorCode.Code), errorCode.Error())

	// Некорректный код — ошибка разработчика
	default:
		log.Error().Str("id", opID).Str("method", method).Err(fmt.Errorf("CodeError.Code is incorrect: %d", errorCode.Code)).Msg("error")
		return status.Errorf(codes.Internal, "internal error")
	}
}
