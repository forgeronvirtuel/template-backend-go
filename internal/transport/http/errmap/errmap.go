package errmap

import (
	"errors"
	"net/http"

	"template-backend-go/internal/domain/apperr"
)

type Mapped struct {
	HTTPStatus int
	Code       string
	Message    string
}

func Map(err error) Mapped {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		switch ae.Code {
		case apperr.CodeNotFound:
			return Mapped{HTTPStatus: http.StatusNotFound, Code: string(ae.Code), Message: ae.Message}
		case apperr.CodeConflict:
			return Mapped{HTTPStatus: http.StatusConflict, Code: string(ae.Code), Message: ae.Message}
		case apperr.CodeUnauthorized:
			return Mapped{HTTPStatus: http.StatusUnauthorized, Code: string(ae.Code), Message: ae.Message}
		case apperr.CodeForbidden:
			return Mapped{HTTPStatus: http.StatusForbidden, Code: string(ae.Code), Message: ae.Message}
		default:
			return Mapped{HTTPStatus: http.StatusInternalServerError, Code: string(apperr.CodeInternal), Message: "Internal server error"}
		}
	}

	// Unknown error type -> internal
	return Mapped{HTTPStatus: http.StatusInternalServerError, Code: string(apperr.CodeInternal), Message: "Internal server error"}
}
