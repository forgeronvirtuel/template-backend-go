package response

import "github.com/gin-gonic/gin"

type Meta struct {
	RequestID string `json:"request_id"`
}

type Success struct {
	Data any  `json:"data"`
	Meta Meta `json:"meta"`
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Rule    string `json:"rule,omitempty"`
	Message string `json:"message"`
}

type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type Failure struct {
	Error ErrorBody `json:"error"`
	Meta  Meta      `json:"meta"`
}

func OK(c *gin.Context, status int, requestID string, data any) {
	c.JSON(status, Success{
		Data: data,
		Meta: Meta{
			RequestID: requestID,
		},
	})
}

func Fail(c *gin.Context, status int, requestID string, code, message string, details []ErrorDetail) {
	c.JSON(status, Failure{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: Meta{
			RequestID: requestID,
		},
	})
}
