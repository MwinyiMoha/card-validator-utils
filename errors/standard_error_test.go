package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
)

type FaultyErrorDetail struct{}

func TestError_ErrorMethod(t *testing.T) {
	err := NewErrorf(NotFound, "Item %s not found", "123")
	assert.NotNil(t, err)
}

func TestNewErrorf(t *testing.T) {
	err := NewErrorf(NotFound, "Item %s not found", "123")
	assert.NotNil(t, err)

	customErr, ok := err.(*Error)
	assert.True(t, ok, "Expected error type *errors.Error")

	expectedMessage := "Item 123 not found"
	assert.Equal(t, expectedMessage, customErr.Message, "Expected error message to be formatted correctly")

	expectedMessage = "code: 2, message: Item 123 not found"
	assert.Equal(t, expectedMessage, customErr.Error())
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("database connection failed")
	err := WrapError(originalErr, Internal, "Could not fetch data")

	assert.NotNil(t, err)

	customErr, ok := err.(*Error)
	assert.True(t, ok, "Expected error type *errors.Error")

	expectedMessage := "Could not fetch data"
	assert.Equal(t, expectedMessage, customErr.Message, "Expected error message")
}

func TestWrapErrorf(t *testing.T) {
	originalErr := errors.New("database connection failed")

	err := WrapErrorf(originalErr, Internal, "Failed to retrieve data")
	assert.NotNil(t, err)

	customErr, ok := err.(*Error)
	assert.True(t, ok, "Expected error type *errors.Error")
	assert.Equal(t, originalErr, customErr.Unwrap(), "Expected unwrapped error to be the original error")
	assert.True(t, errors.Is(customErr, originalErr), "Expected original error to be wrapped")

	expectedMessage := "Failed to retrieve data"
	assert.Equal(t, expectedMessage, customErr.Message, "Expected error message")
}

func TestErrorFormatting(t *testing.T) {
	originalErr := errors.New("underlying system failure")
	err := WrapErrorf(originalErr, Internal, "Critical operation failed")

	expected := "code: 4, message: Critical operation failed, original error: underlying system failure"
	assert.Equal(t, expected, err.Error(), "Expected error string to be formatted correctly")
}

func TestUnwrap(t *testing.T) {
	originalErr := errors.New("network timeout")
	err := WrapErrorf(originalErr, Internal, "Request timeout")

	unwrapped := errors.Unwrap(err)
	assert.Equal(t, originalErr, unwrapped, "Expected unwrapped error to be the original error")
}

func TestErrorCodeCheck(t *testing.T) {
	err := NewErrorf(Unauthorized, "Unauthorized access")

	customErr, ok := err.(*Error)
	assert.True(t, ok, "Expected error type *errors.Error")
	assert.Equal(t, Unauthorized, customErr.Code(), "Expected error code to be Unauthorized")
}

func TestMapErrorCodeToGRPCCode(t *testing.T) {
	assert.Equal(t, codes.Internal, mapErrorCodeToGRPCCode(Internal))
	assert.Equal(t, codes.InvalidArgument, mapErrorCodeToGRPCCode(BadRequest))
	assert.Equal(t, codes.ResourceExhausted, mapErrorCodeToGRPCCode(QuotaExceeded))
	assert.Equal(t, codes.NotFound, mapErrorCodeToGRPCCode(NotFound))
	assert.Equal(t, codes.Unauthenticated, mapErrorCodeToGRPCCode(Unauthenticated))
	assert.Equal(t, codes.PermissionDenied, mapErrorCodeToGRPCCode(Unauthorized))
	assert.Equal(t, codes.AlreadyExists, mapErrorCodeToGRPCCode(Conflict))
	assert.Equal(t, codes.Unknown, mapErrorCodeToGRPCCode(Unknown))
}

func TestGRPCStatus(t *testing.T) {
	originalErr := errors.New("database connection failed")
	customErr := WrapErrorf(originalErr, Internal, "Failed to retrieve data")

	grpcStatus := customErr.(*Error).GRPCStatus()
	assert.Equal(t, codes.Internal, grpcStatus.Code(), "Expected gRPC code to be Internal")

	expectedMessage := "Failed to retrieve data"
	assert.Equal(t, expectedMessage, grpcStatus.Message(), "Expected gRPC message to be correct")
}

func TestGRPCStatus_NoOriginalError(t *testing.T) {
	customErr := NewErrorf(Internal, "database connection failed")
	grpcStatus := customErr.(*Error).GRPCStatus()

	assert.Equal(t, codes.Internal, grpcStatus.Code(), "Expected gRPC code to be Internal")

	expectedMessage := "database connection failed"
	assert.Equal(t, expectedMessage, grpcStatus.Message(), "Expected gRPC message to be correct")
}

func TestGRPCStatus_WithErrorDetails(t *testing.T) {
	originalErr := errors.New("database connection failed")
	customErr := WrapErrorf(originalErr, Internal, "Failed to retrieve data")
	grpcStatus := customErr.(*Error).GRPCStatus()

	for _, detail := range grpcStatus.Details() {
		switch info := detail.(type) {
		case *errdetails.ErrorInfo:
			assert.Equal(t, "database connection failed", info.Reason, "Expected reason to match the original error")
			assert.Equal(t, "4", info.Metadata["custom_code"], "Expected custom code metadata")
		}
	}
}
