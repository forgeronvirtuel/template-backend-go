package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"
const RequestIDHeader = "X-Request-Id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(RequestIDHeader)
		if rid == "" {
			rid = newRequestID()
		}

		c.Set(RequestIDKey, rid)
		c.Writer.Header().Set(RequestIDHeader, rid)

		c.Next()
	}
}

func newRequestID() (string, []error) {
	var b [16]byte
	errors := []error{}
	for i := 0; i < 3; i++ {
		_, err := rand.Read(b[:])
		if err != nil {
			errors = append(errors, err)
		} else {
			break
		}
	}
	if len(errors) > 0 {
		return "", errors
	}
	return hex.EncodeToString(b[:]), nil
}
