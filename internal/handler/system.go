package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"homework/internal/model"
)

// Health відповідає на GET /health. Не належить жодному ресурсу.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// NotFound відповідає на кожен незареєстрований маршрут у форматі JSON-помилки.
func NotFound(c *gin.Context) {
	respondError(c, http.StatusNotFound, model.CodeNotFound, "Route not found")
}
