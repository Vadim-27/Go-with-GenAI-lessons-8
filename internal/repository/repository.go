package repository

import (
	"errors"

	"homework/internal/model"
)

// ErrNotFound is returned when a book with the requested id does not exist.
var ErrNotFound = errors.New("book not found")

// BookRepository is the storage contract used by the handlers. Handlers
// depend on this interface, not on a concrete implementation, so the
// in-memory store can later be replaced (for example by a database).
type BookRepository interface {
	// Create stores a new book. The repository assigns ID, CreatedAt and
	// UpdatedAt; any values in the argument are ignored.
	Create(book model.Book) model.Book

	// Get returns the book with the given id or ErrNotFound.
	Get(id uint32) (model.Book, error)

	// Update fully replaces the book with the given id. ID and CreatedAt are
	// kept, UpdatedAt is refreshed. Returns ErrNotFound if it does not exist.
	Update(id uint32, book model.Book) (model.Book, error)

	// Delete removes the book or returns ErrNotFound.
	Delete(id uint32) error

	// List returns books ordered by id ascending. A non-empty category keeps
	// only books with exactly that category. page and limit must be >= 1.
	List(category string, page, limit int) []model.Book
}
