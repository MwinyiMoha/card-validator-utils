package errors

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidationError_ErrorMethod(t *testing.T) {
	violations := []*FieldViolation{
		{Field: "Name", Description: "Name is required"},
		{Field: "Age", Description: "Age must be greater than 0"},
	}
	validationErr := NewValidationError(violations)

	errMsg := validationErr.Error()

	assert.Contains(t, errMsg, "validation error(s):")
	assert.Contains(t, errMsg, "Name: Name is required")
	assert.Contains(t, errMsg, "Age: Age must be greater than 0")
}

func TestValidationError_GRPCStatus(t *testing.T) {
	violations := []*FieldViolation{
		{Field: "Name", Description: "Name is required"},
		{Field: "Age", Description: "Age must be greater than 0"},
	}
	validationErr := NewValidationError(violations)

	st := validationErr.GRPCStatus()
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "invalid request", st.Message())

	details := st.Details()
	assert.Len(t, details, 1)

	badRequest, ok := details[0].(*errdetails.BadRequest)
	assert.True(t, ok)
	assert.Len(t, badRequest.FieldViolations, 2)

	assert.Equal(t, "Name", badRequest.FieldViolations[0].Field)
	assert.Equal(t, "Name is required", badRequest.FieldViolations[0].Description)
	assert.Equal(t, "Age", badRequest.FieldViolations[1].Field)
	assert.Equal(t, "Age must be greater than 0", badRequest.FieldViolations[1].Description)
}

func TestBuildViolations(t *testing.T) {
	validate := validator.New()
	type Payload struct {
		Name string `validate:"required"`
		Age  int    `validate:"gte=1"`
	}

	payload := &Payload{
		Name: "", // Invalid: Name is required
		Age:  0,  // Invalid: Age must be greater than or equal to 1
	}

	err := validate.Struct(payload)
	assert.Error(t, err)

	verrs := err.(validator.ValidationErrors)
	violations := BuildViolations(verrs)

	assert.Len(t, violations, 2)
	assert.Equal(t, "Name", violations[0].Field)
	assert.Contains(t, violations[0].Description, "required")
	assert.Equal(t, "Age", violations[1].Field)
	assert.Contains(t, violations[1].Description, "gte")
}

func TestValidationError_GRPCStatus_NoViolations(t *testing.T) {
	validationErr := NewValidationError(nil)

	st := validationErr.GRPCStatus()

	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Equal(t, "invalid request", st.Message())

	details := st.Details()
	assert.Len(t, details, 1)

	badRequest, ok := details[0].(*errdetails.BadRequest)
	assert.True(t, ok)
	assert.Empty(t, badRequest.FieldViolations)
}

func TestNewValidationError_EmptyViolations(t *testing.T) {
	validationErr := NewValidationError(nil)

	assert.NotNil(t, validationErr)
	assert.Empty(t, validationErr.FieldViolations)
}

func TestBuildViolations_Empty(t *testing.T) {
	var verrs validator.ValidationErrors
	violations := BuildViolations(verrs)

	assert.Empty(t, violations)
}

func TestValidationError_GRPCStatus_NonValidationError(t *testing.T) {
	nonValidationErr := errors.New("some internal error")
	st := status.New(codes.Internal, nonValidationErr.Error())

	grpcErr := st.Err()

	assert.Equal(t, codes.Internal, status.Code(grpcErr))
	assert.Contains(t, grpcErr.Error(), "some internal error")
}
