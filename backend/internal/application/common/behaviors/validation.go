package behaviors

import (
	"context"
	"fmt"

	"kuorum/backend/internal/application/common/mediator"
)

// Validatable represents any Command or Query that provides validation rules.
type Validatable interface {
	Validate() error
}

// ValidationError wraps validation failure details.
type ValidationError struct {
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (v *ValidationError) Error() string {
	if v.Err != nil {
		return fmt.Sprintf("%s: %v", v.Message, v.Err)
	}
	return v.Message
}

func (v *ValidationError) Unwrap() error {
	return v.Err
}

// ValidationBehavior inspects requests for the Validatable interface and returns ValidationError on failure.
func ValidationBehavior() mediator.PipelineBehavior {
	return func(ctx context.Context, req any, next func(ctx context.Context, req any) (any, error)) (any, error) {
		if v, ok := req.(Validatable); ok {
			if err := v.Validate(); err != nil {
				return nil, &ValidationError{
					Message: "Validation failed",
					Err:     err,
				}
			}
		}
		return next(ctx, req)
	}
}
