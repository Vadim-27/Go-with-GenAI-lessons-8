// Package app wires the HTTP API together.
package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"libratrack/internal/handler"
	"libratrack/internal/middleware"
	"libratrack/internal/repository"
)

func init() {
	// Release mode: no debug banner or route table in the logs.
	gin.SetMode(gin.ReleaseMode)
}

// NewRouter builds the router: it creates the repositories on top of db,
// the handlers on top of the repositories, and registers all routes.
func NewRouter(db *gorm.DB) http.Handler {
	categories := handler.NewCategoryHandler(repository.NewCategoryRepository(db))
	authors := handler.NewAuthorHandler(repository.NewAuthorRepository(db))
	books := handler.NewBookHandler(repository.NewBookRepository(db))
	copies := handler.NewCopyHandler(repository.NewCopyRepository(db))

	r := gin.New()
	// Order matters: Logger is the outermost so it sees the 500 after a panic;
	// Recovery comes before CORS; CORS sets headers before the handler runs.
	r.Use(middleware.Logger(), middleware.Recovery(), middleware.CORS())
	// A trailing slash is a different, unregistered route (JSON 404), not a 307 redirect.
	r.RedirectTrailingSlash = false
	r.NoRoute(handler.NotFound)
	r.GET("/health", handler.Health)
	r.GET("/openapi.yaml", handler.OpenAPI)
	r.GET("/swagger", handler.SwaggerUI)
	r.GET("/swagger/index.html", handler.SwaggerUI)

	api := r.Group("/api/v1")
	categories.Register(api)
	authors.Register(api)
	books.Register(api)
	copies.Register(api)

	return r
}
