package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

// CategoryHandler serves the /categories resource.
type CategoryHandler struct {
	repo repository.CategoryRepository
}

// NewCategoryHandler returns a handler backed by repo.
func NewCategoryHandler(repo repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

// Register attaches the category routes to the group.
func (h *CategoryHandler) Register(r *gin.RouterGroup) {
	r.POST("/categories", h.CreateCategory)
	r.GET("/categories", h.ListCategories)
	r.GET("/categories/:id", h.GetCategory)
}

// CreateCategory answers POST /categories.
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var in model.CategoryInput
	if !bindInput(c, &in) {
		return
	}

	category := in.ToCategory()
	if err := h.repo.Create(c.Request.Context(), &category); err != nil {
		respondRepoError(c, err, "Category")
		return
	}
	c.JSON(http.StatusCreated, category)
}

// ListCategories answers GET /categories?page=&limit=.
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	page, limit, ok := pagination(c)
	if !ok {
		return
	}

	categories, err := h.repo.List(c.Request.Context(), page, limit)
	if err != nil {
		respondRepoError(c, err, "Category")
		return
	}
	c.JSON(http.StatusOK, categories)
}

// GetCategory answers GET /categories/{id}.
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	category, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, err, "Category")
		return
	}
	c.JSON(http.StatusOK, category)
}
