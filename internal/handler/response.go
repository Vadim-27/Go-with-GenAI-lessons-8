package handler

import (
	"github.com/gin-gonic/gin"

	"homework/internal/model"
)

// respondError пише JSON-обгортку помилки із заданим HTTP-статусом.
func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, model.NewErrorResponse(code, message))
}
