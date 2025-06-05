package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"strings"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // восстановить тело запроса
		}

		var headers []string
		for name, values := range c.Request.Header {
			for _, value := range values {
				headers = append(headers, name+": "+value)
			}
		}

		log.Printf("---- Incoming Request ----\n"+
			"Method: %s\nPath: %s\nHeaders:\n%s\nBody:\n%s\n--------------------------",
			c.Request.Method,
			c.Request.URL.Path,
			strings.Join(headers, "\n"),
			string(bodyBytes),
		)

		c.Next()
	}
}
