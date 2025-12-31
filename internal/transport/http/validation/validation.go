package validation

import (
	"errors"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"template-backend-go/internal/transport/http/response"
)

func FromBindError(err error) []response.ErrorDetail {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		out := make([]response.ErrorDetail, 0, len(verrs))
		for _, fe := range verrs {
			out = append(out, response.ErrorDetail{
				Field:   fe.Field(),
				Rule:    fe.Tag(),
				Message: humanMessage(fe),
			})
		}
		return out
	}

	// For JSON parse errors, type errors, etc.
	return []response.ErrorDetail{
		{Message: err.Error()},
	}
}

func humanMessage(fe validator.FieldError) string {
	// Keep it simple; can be improved later.
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "is too short"
	case "max":
		return "is too long"
	default:
		return "is invalid"
	}
}

// Optional: ensure Gin uses the validator engine you expect.
func EnsureValidatorInitialized() {
	_ = binding.Validator.Engine()
}
