package errors

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FieldViolation struct {
	Field       string
	Description string
}

type ValidationError struct {
	FieldViolations []*FieldViolation
}

func (e *ValidationError) Error() string {
	var violationMessages []string
	for _, v := range e.FieldViolations {
		violationMessages = append(violationMessages, fmt.Sprintf("%s: %s", v.Field, v.Description))
	}

	return fmt.Sprintf("validation error(s): %v", violationMessages)
}

func (e *ValidationError) GRPCStatus() (*status.Status, error) {
	st := status.New(codes.InvalidArgument, "invalid request")

	violations := []*errdetails.BadRequest_FieldViolation{}
	for _, v := range e.FieldViolations {
		violations = append(
			violations,
			&errdetails.BadRequest_FieldViolation{
				Field:       v.Field,
				Description: v.Description,
			},
		)
	}

	return st.WithDetails(
		&errdetails.BadRequest{
			FieldViolations: violations,
		},
	)
}

func NewValidationError(violations []*FieldViolation) *ValidationError {
	return &ValidationError{FieldViolations: violations}
}

func BuildViolations(verrs validator.ValidationErrors) []*FieldViolation {
	var violations []*FieldViolation

	for _, err := range verrs {
		violations = append(
			violations,
			&FieldViolation{
				Field:       err.StructField(),
				Description: err.Error(),
			},
		)
	}

	return violations
}
