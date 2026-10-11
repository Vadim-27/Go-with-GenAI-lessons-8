// Package app wires the whole HTTP API together.
//
// The automatic tests talk to your API ONLY through NewRouter(), so you are free to
// organise the rest of the code (model, repository, handler, middleware) as you like.
package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"homework/internal/handler"
	"homework/internal/middleware"
	"homework/internal/repository"
)

func init() {
	// Вимикаємо debug-режим Gin: без банера та таблиці маршрутів у логах.
	gin.SetMode(gin.ReleaseMode)
}

// NewRouter повертає повністю налаштований HTTP-обробник ресурсу books.
//
// Кожен виклик будує НОВИЙ роутер зі своїм ПОРОЖНІМ сховищем у пам'яті,
// тому роутери ніколи не ділять стан між собою.
func NewRouter() http.Handler {
	repo := repository.NewMemoryRepository()
	books := handler.NewBookHandler(repo)

	r := gin.New()
	// Порядок важливий: Logger ззовні, щоб бачити і 500 після паніки;
	// Recovery перед CORS, а CORS ставить заголовки ще до хендлера.
	r.Use(middleware.Logger(), middleware.Recovery(), middleware.CORS())
	// Шлях зі слешем наприкінці — це інший, незареєстрований маршрут (404 JSON), а не редирект 307.
	r.RedirectTrailingSlash = false
	r.NoRoute(handler.NotFound)
	r.GET("/health", handler.Health)
	r.GET("/openapi.yaml", handler.OpenAPI)
	r.GET("/swagger", handler.SwaggerUI)

	// Новий ресурс — це ще один виклик Register тут.
	api := r.Group("/api/v1")
	books.Register(api)

	return r
}
