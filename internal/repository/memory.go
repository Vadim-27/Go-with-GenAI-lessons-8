package repository

import (
	"math"
	"slices"
	"sync"
	"time"

	"homework/internal/model"
)

// MemoryRepository is a BookRepository that keeps everything in memory.
// It is safe for concurrent use.
type MemoryRepository struct {
	mu     sync.RWMutex
	books  map[uint32]model.Book
	nextID uint32
}

// NewMemoryRepository returns an empty store. Ids start at 1.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		books:  make(map[uint32]model.Book),
		nextID: 1,
	}
}

func (r *MemoryRepository) Create(book model.Book) model.Book {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	book = clone(book)
	book.ID = r.nextID
	book.CreatedAt = now
	book.UpdatedAt = now

	// nextID only grows, so an id is never reused after a delete.
	r.nextID++
	r.books[book.ID] = book
	return clone(book)
}

func (r *MemoryRepository) Get(id uint32) (model.Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	book, ok := r.books[id]
	if !ok {
		return model.Book{}, ErrNotFound
	}
	return clone(book), nil
}

func (r *MemoryRepository) Update(id uint32, book model.Book) (model.Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	old, ok := r.books[id]
	if !ok {
		return model.Book{}, ErrNotFound
	}

	book = clone(book)
	book.ID = old.ID
	book.CreatedAt = old.CreatedAt
	book.UpdatedAt = time.Now().UTC()
	if book.UpdatedAt.Before(old.UpdatedAt) {
		book.UpdatedAt = old.UpdatedAt
	}

	r.books[id] = book
	return clone(book), nil
}

func (r *MemoryRepository) Delete(id uint32) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.books[id]; !ok {
		return ErrNotFound
	}
	delete(r.books, id)
	return nil
}

func (r *MemoryRepository) List(category string, page, limit int) []model.Book {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Map iteration order is random, so collect the ids and sort them.
	ids := make([]uint32, 0, len(r.books))
	for id, book := range r.books {
		if category == "" || book.Category == category {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)

	// (page-1)*limit overflows int for huge pages, so compare before multiplying.
	if page-1 > math.MaxInt/limit {
		return []model.Book{}
	}
	start := (page - 1) * limit
	if start >= len(ids) {
		return []model.Book{}
	}
	end := start + limit
	if end > len(ids) {
		end = len(ids)
	}

	// A non-nil slice marshals to [] and not to null.
	result := make([]model.Book, 0, end-start)
	for _, id := range ids[start:end] {
		result = append(result, clone(r.books[id]))
	}
	return result
}

// clone makes a deep copy of a book. Book holds pointers (PublishedYear,
// Description); copying only the struct would share them with the caller and
// allow data races or changes of stored data from outside.
func clone(b model.Book) model.Book {
	if b.PublishedYear != nil {
		year := *b.PublishedYear
		b.PublishedYear = &year
	}
	if b.Description != nil {
		desc := *b.Description
		b.Description = &desc
	}
	return b
}
