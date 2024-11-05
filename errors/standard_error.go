package errors

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorCode int

const (
	Unknown ErrorCode = iota + 1
	NotFound
	BadRequest
	Internal
	Unauthenticated
	Unauthorized
	Conflict
	QuotaExceeded
)

type Error struct {
	Original error     `json:"original_error"`
	Message  string    `json:"message"`
	ErrCode  ErrorCode `json:"error_code"`
}

func (e *Error) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("code: %d, message: %s, original error: %s", e.ErrCode, e.Message, e.Original)
	}

	return fmt.Sprintf("code: %d, message: %s", e.ErrCode, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Original
}

func (e *Error) Code() ErrorCode {
	return e.ErrCode
}

func (e *Error) GRPCStatus() *status.Status {
	errCode := mapErrorCodeToGRPCCode(e.ErrCode)
	st := status.New(errCode, e.Message)

	if e.Original != nil {
		details := &errdetails.ErrorInfo{
			Reason:   e.Original.Error(),
			Metadata: map[string]string{"custom_code": fmt.Sprintf("%d", e.ErrCode)},
		}

		detailed, err := st.WithDetails(details)
		if err != nil {
			return st
		}

		return detailed
	}

	return st
}

func WrapError(orig error, code ErrorCode, message string) error {
	return &Error{
		ErrCode:  code,
		Original: orig,
		Message:  message,
	}
}

func WrapErrorf(orig error, code ErrorCode, format string, a ...interface{}) error {
	return &Error{
		ErrCode:  code,
		Original: orig,
		Message:  fmt.Sprintf(format, a...),
	}
}

func NewErrorf(code ErrorCode, format string, a ...interface{}) error {
	return WrapErrorf(nil, code, format, a...)
}

func mapErrorCodeToGRPCCode(code ErrorCode) codes.Code {
	switch code {
	case NotFound:
		return codes.NotFound
	case BadRequest:
		return codes.InvalidArgument
	case Internal:
		return codes.Internal
	case Unauthenticated:
		return codes.Unauthenticated
	case Unauthorized:
		return codes.PermissionDenied
	case Conflict:
		return codes.AlreadyExists
	case QuotaExceeded:
		return codes.ResourceExhausted
	default:
		return codes.Unknown
	}
}
