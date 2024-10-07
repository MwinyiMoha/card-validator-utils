package errors

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoiface"
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
	st := status.New(mapErrorCodeToGRPCCode(e.ErrCode), e.Message)

	if e.Original != nil {
		details := &errdetails.ErrorInfo{
			Reason:   e.Original.Error(),
			Metadata: map[string]string{"custom_code": fmt.Sprintf("%d", e.ErrCode)},
		}

		return grpcStatusWithDetails(st, details)
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

func grpcStatusWithDetails(st *status.Status, details ...protoiface.MessageV1) *status.Status {
	for _, detail := range details {
		st, err := st.WithDetails(detail)
		if err != nil {
			return st
		}
	}
	return st
}
