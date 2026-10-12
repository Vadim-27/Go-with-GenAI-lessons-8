package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"libratrack/internal/model"
	"libratrack/internal/repository"
)

// AuthorHandler serves the /authors resource.
type AuthorHandler struct {
	repo repository.AuthorRepository
}

// NewAuthorHandler returns a handler backed by repo.
func NewAuthorHandler(repo repository.AuthorRepository) *AuthorHandler {
	return &AuthorHandler{repo: repo}
}

// Register attaches the author routes to the group.
func (h *AuthorHandler) Register(r *gin.RouterGroup) {
	r.POST("/authors", h.CreateAuthor)
	r.GET("/authors", h.ListAuthors)
	r.GET("/authors/:id", h.GetAuthor)
}

// CreateAuthor answers POST /authors.
func (h *AuthorHandler) CreateAuthor(c *gin.Context) {
	var in model.AuthorInput
	if !bindInput(c, &in) {
		return
	}

	author := in.ToAuthor()
	if err := h.repo.Create(c.Request.Context(), &author); err != nil {
		respondRepoError(c, err, "Author")
		return
	}
	c.JSON(http.StatusCreated, author)
}

// ListAuthors answers GET /authors?page=&limit=.
func (h *AuthorHandler) ListAuthors(c *gin.Context) {
	page, limit, ok := pagination(c)
	if !ok {
		return
	}

	authors, err := h.repo.List(c.Request.Context(), page, limit)
	if err != nil {
		respondRepoError(c, err, "Author")
		return
	}
	c.JSON(http.StatusOK, authors)
}

// GetAuthor answers GET /authors/{id}.
func (h *AuthorHandler) GetAuthor(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	author, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, err, "Author")
		return
	}
	c.JSON(http.StatusOK, author)
}
