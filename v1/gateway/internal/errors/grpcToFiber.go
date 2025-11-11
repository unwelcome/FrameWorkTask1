package Error

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCErrorToHTTP(err error, c *fiber.Ctx) error {
	if err == nil {
		return c.Status(fiber.StatusOK).JSON(HttpError{Code: 200, Message: ""})
	}

	// Если это не gRPC ошибка, возвращаем 500
	st, ok := status.FromError(err)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(HttpError{Code: 500, Message: err.Error()})
	}

	switch st.Code() {
	case codes.OK:
		return c.Status(fiber.StatusOK).JSON(HttpError{Code: 200, Message: ""})
	case codes.Canceled:
		return c.Status(fiber.StatusRequestTimeout).JSON(HttpError{Code: 408, Message: st.Message()})
	case codes.Unknown:
		return c.Status(fiber.StatusInternalServerError).JSON(HttpError{Code: 500, Message: st.Message()})
	case codes.InvalidArgument:
		return c.Status(fiber.StatusBadRequest).JSON(HttpError{Code: 400, Message: st.Message()})
	case codes.DeadlineExceeded:
		return c.Status(fiber.StatusGatewayTimeout).JSON(HttpError{Code: 504, Message: st.Message()})
	case codes.NotFound:
		return c.Status(fiber.StatusNotFound).JSON(HttpError{Code: 404, Message: st.Message()})
	case codes.AlreadyExists:
		return c.Status(fiber.StatusConflict).JSON(HttpError{Code: 409, Message: st.Message()})
	case codes.PermissionDenied:
		return c.Status(fiber.StatusForbidden).JSON(HttpError{Code: 403, Message: st.Message()})
	case codes.ResourceExhausted:
		return c.Status(fiber.StatusTooManyRequests).JSON(HttpError{Code: 429, Message: st.Message()})
	case codes.FailedPrecondition:
		return c.Status(fiber.StatusPreconditionFailed).JSON(HttpError{Code: 412, Message: st.Message()})
	case codes.Aborted:
		return c.Status(fiber.StatusConflict).JSON(HttpError{Code: 409, Message: st.Message()})
	case codes.OutOfRange:
		return c.Status(fiber.StatusBadRequest).JSON(HttpError{Code: 400, Message: st.Message()})
	case codes.Unimplemented:
		return c.Status(fiber.StatusNotImplemented).JSON(HttpError{Code: 501, Message: st.Message()})
	case codes.Internal:
		return c.Status(fiber.StatusInternalServerError).JSON(HttpError{Code: 500, Message: st.Message()})
	case codes.Unavailable:
		return c.Status(fiber.StatusServiceUnavailable).JSON(HttpError{Code: 503, Message: st.Message()})
	case codes.DataLoss:
		return c.Status(fiber.StatusInternalServerError).JSON(HttpError{Code: 500, Message: st.Message()})
	case codes.Unauthenticated:
		return c.Status(fiber.StatusUnauthorized).JSON(HttpError{Code: 401, Message: st.Message()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(HttpError{Code: 500, Message: st.Message()})
	}
}
