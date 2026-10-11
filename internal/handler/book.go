package handler

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"homework/internal/model"
	"homework/internal/repository"
)

// Параметри пагінації списку.
const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

// BookHandler містить HTTP-обробники ресурсу books.
// Залежить від інтерфейсу сховища, а не від конкретної реалізації.
type BookHandler struct {
	repo repository.BookRepository
}

// NewBookHandler повертає хендлер, що зберігає дані в repo.
func NewBookHandler(repo repository.BookRepository) *BookHandler {
	return &BookHandler{repo: repo}
}

// Register підключає маршрути books до переданої групи. Кожен хендлер
// володіє своїми маршрутами, тому app.go не росте з кожним новим ресурсом.
func (h *BookHandler) Register(r *gin.RouterGroup) {
	r.POST("/books", h.CreateBook)
	r.GET("/books", h.ListBooks)
	r.GET("/books/:id", h.GetBook)
	r.PUT("/books/:id", h.UpdateBook)
	r.DELETE("/books/:id", h.DeleteBook)
}

// CreateBook відповідає на POST /api/v1/books.
func (h *BookHandler) CreateBook(c *gin.Context) {
	book, ok := bindBook(c)
	if !ok {
		return
	}

	c.JSON(http.StatusCreated, h.repo.Create(book))
}

// ListBooks відповідає на GET /api/v1/books?page=1&limit=10&category=go.
// Порожня категорія означає «без фільтра».
func (h *BookHandler) ListBooks(c *gin.Context) {
	page, ok := queryPositiveInt(c, "page", defaultPage)
	if !ok {
		respondError(c, http.StatusBadRequest, model.CodeInvalidPagination, "Invalid page")
		return
	}
	limit, ok := queryPositiveInt(c, "limit", defaultLimit)
	if !ok {
		respondError(c, http.StatusBadRequest, model.CodeInvalidPagination, "Invalid limit")
		return
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	c.JSON(http.StatusOK, h.repo.List(c.Query("category"), page, limit))
}

// GetBook відповідає на GET /api/v1/books/{id}.
func (h *BookHandler) GetBook(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	book, err := h.repo.Get(id)
	if err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, book)
}

// UpdateBook відповідає на PUT /api/v1/books/{id}: повна заміна книжки.
// Порядок перевірок: спочатку id (400), потім тіло (422), потім існування (404).
func (h *BookHandler) UpdateBook(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	book, ok := bindBook(c)
	if !ok {
		return
	}

	updated, err := h.repo.Update(id, book)
	if err != nil {
		respondRepoError(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// DeleteBook відповідає на DELETE /api/v1/books/{id}: 204 без тіла.
func (h *BookHandler) DeleteBook(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.repo.Delete(id); err != nil {
		respondRepoError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// bindBook розбирає й валідує тіло запиту. Якщо повертає false, відповідь
// 422 уже записана, і викликач має зупинитись.
func bindBook(c *gin.Context) (model.Book, bool) {
	var in model.BookInput
	if err := decodeBody(c, &in); err != nil {
		// Помилка декодера — внутрішня деталь (типи полів, позиції); клієнту її не показуємо.
		respondError(c, http.StatusUnprocessableEntity, model.CodeValidationError, "Invalid request body")
		return model.Book{}, false
	}
	if err := in.Validate(); err != nil {
		respondError(c, http.StatusUnprocessableEntity, model.CodeValidationError, err.Error())
		return model.Book{}, false
	}

	// Validate гарантує, що обов'язкові вказівники не nil.
	return model.Book{
		Title:         *in.Title,
		ISBN:          *in.ISBN,
		Author:        *in.Author,
		Category:      *in.Category,
		PublishedYear: in.PublishedYear,
		Description:   in.Description,
	}, true
}

// decodeBody розбирає тіло запиту в dst і вимагає, щоб після JSON-об'єкта
// більше нічого не було: `{...} сміття` не є коректним JSON.
func decodeBody(c *gin.Context, dst any) error {
	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return errors.New("unexpected data after JSON body")
	}
	return nil
}

// queryPositiveInt читає цілий параметр запиту >= 1. Якщо параметра немає,
// повертає def. Для нецілого, нульового чи від'ємного значення повертає false.
// Надто велике додатне число обмежується math.MaxInt, щоб не давати 500.
func queryPositiveInt(c *gin.Context, key string, def int) (int, bool) {
	raw, present := c.GetQuery(key)
	if !present {
		return def, true
	}

	n, err := strconv.ParseInt(raw, 10, 64)
	if errors.Is(err, strconv.ErrRange) && !strings.HasPrefix(raw, "-") {
		return math.MaxInt, true
	}
	if err != nil || n < 1 {
		return 0, false
	}
	if n > math.MaxInt {
		return math.MaxInt, true
	}
	return int(n), true
}

// parseID читає параметр шляху {id}. Приймає лише 0..4294967295; для
// будь-чого іншого відповідь 400 уже записана.
func parseID(c *gin.Context) (uint32, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, model.CodeInvalidID, "Invalid book id")
		return 0, false
	}
	return uint32(id), true
}

// respondRepoError перетворює помилку сховища на HTTP-відповідь.
func respondRepoError(c *gin.Context, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		respondError(c, http.StatusNotFound, model.CodeNotFound, "Book not found")
		return
	}
	// Невідома помилка: деталі йдуть у лог, клієнту про них не кажемо.
	_ = c.Error(err)
	respondError(c, http.StatusInternalServerError, model.CodeInternalServerError, "Internal server error")
}
