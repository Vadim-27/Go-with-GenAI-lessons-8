package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger пише в лог один рядок на запит: метод, шлях, статус і тривалість.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		log.Printf("%s %q -> %d (%s)",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}
