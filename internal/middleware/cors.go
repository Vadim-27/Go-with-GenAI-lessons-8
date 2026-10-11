package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS додає до кожної відповіді заголовок Access-Control-Allow-Origin і
// сам відповідає на preflight-запити (OPTIONS) статусом 204.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", "*")

		if c.Request.Method != http.MethodOptions {
			c.Next()
			return
		}

		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Дозволяємо ті заголовки, які браузер просить у preflight-запиті.
		if reqHeaders := c.GetHeader("Access-Control-Request-Headers"); reqHeaders != "" {
			h.Set("Access-Control-Allow-Headers", reqHeaders)
		} else {
			h.Set("Access-Control-Allow-Headers", "Content-Type")
		}
		h.Set("Access-Control-Max-Age", "86400")

		// Preflight завершується тут: до хендлерів і маршрутів він не доходить.
		c.AbortWithStatus(http.StatusNoContent)
	}
}
