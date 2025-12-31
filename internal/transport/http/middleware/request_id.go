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

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
