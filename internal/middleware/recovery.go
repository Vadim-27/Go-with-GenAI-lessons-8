package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"homework/internal/model"
)

// Recovery перехоплює паніку в обробнику. Деталі (текст паніки і stack trace)
// йдуть лише в лог, а клієнт отримує 500 internal_server_error без подробиць.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			// Це штатний спосіб net/http перервати відповідь, його не глушимо.
			if rec == http.ErrAbortHandler {
				panic(rec)
			}

			log.Printf("panic recovered: %v\n%s", rec, debug.Stack())

			if c.Writer.Written() {
				// Відповідь уже почала відправлятись, змінити статус не можна.
				c.Abort()
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				model.NewErrorResponse(model.CodeInternalServerError, "Internal server error"))
		}()

		c.Next()
	}
}
