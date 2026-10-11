package repository_test

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sync"
	"testing"
	"time"

	"homework/internal/model"
	"homework/internal/repository"
)

func ptr[T any](v T) *T { return &v }

func newBook(title, category string) model.Book {
	return model.Book{
		Title:    title,
		ISBN:     "isbn-" + title,
		Author:   "author",
		Category: category,
	}
}

func ids(books []model.Book) []uint32 {
	out := make([]uint32, 0, len(books))
	for _, b := range books {
		out = append(out, b.ID)
	}
	return out
}

func TestCreate_AssignsIDAndTimestamps(t *testing.T) {
	repo := repository.NewMemoryRepository()

	// ID and timestamps supplied by the caller must be ignored.
	got := repo.Create(model.Book{
		ID:        99,
		Title:     "t",
		CreatedAt: time.Unix(0, 0),
		UpdatedAt: time.Unix(0, 0),
	})

	if got.ID != 1 {
		t.Errorf("first id = %d, want 1", got.ID)
	}
	if got.CreatedAt.IsZero() || got.CreatedAt.Unix() == 0 {
		t.Errorf("CreatedAt was not set by the repository: %v", got.CreatedAt)
	}
	if !got.CreatedAt.Equal(got.UpdatedAt) {
		t.Errorf("CreatedAt %v and UpdatedAt %v must be equal on create", got.CreatedAt, got.UpdatedAt)
	}
	if got.CreatedAt.Location() != time.UTC {
		t.Errorf("CreatedAt location = %v, want UTC", got.CreatedAt.Location())
	}
}

func TestCreate_IDsIncreaseAndAreNotReused(t *testing.T) {
	repo := repository.NewMemoryRepository()

	first := repo.Create(newBook("a", "go"))
	second := repo.Create(newBook("b", "go"))
	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("ids = %d, %d, want 1, 2", first.ID, second.ID)
	}

	if err := repo.Delete(second.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	third := repo.Create(newBook("c", "go"))
	if third.ID != 3 {
		t.Errorf("id after delete = %d, want 3 (ids must not be reused)", third.ID)
	}
}

func TestGet(t *testing.T) {
	repo := repository.NewMemoryRepository()
	created := repo.Create(newBook("a", "go"))

	tests := []struct {
		name    string
		id      uint32
		wantErr error
	}{
		{"existing", created.ID, nil},
		{"zero id", 0, repository.ErrNotFound},
		{"unknown id", 42, repository.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.Get(tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got.ID != tt.id {
				t.Errorf("got id %d, want %d", got.ID, tt.id)
			}
		})
	}
}

func TestUpdate_ReplacesFieldsAndKeepsIdentity(t *testing.T) {
	repo := repository.NewMemoryRepository()
	orig := newBook("old", "go")
	orig.PublishedYear = ptr(2015)
	orig.Description = ptr("old description")
	created := repo.Create(orig)

	// The replacement has no optional fields: they must become nil.
	got, err := repo.Update(created.ID, newBook("new", "python"))
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("id = %d, want %d", got.ID, created.ID)
	}
	if !got.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf("CreatedAt changed: %v -> %v", created.CreatedAt, got.CreatedAt)
	}
	if got.UpdatedAt.Before(created.UpdatedAt) {
		t.Errorf("UpdatedAt went backwards: %v -> %v", created.UpdatedAt, got.UpdatedAt)
	}
	if got.Title != "new" || got.Category != "python" {
		t.Errorf("required fields were not replaced: %+v", got)
	}
	if got.PublishedYear != nil || got.Description != nil {
		t.Errorf("optional fields must be cleared, got %+v", got)
	}

	stored, err := repo.Get(created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if stored.Title != "new" || stored.PublishedYear != nil {
		t.Errorf("update was not stored: %+v", stored)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := repository.NewMemoryRepository()

	_, err := repo.Update(7, newBook("a", "go"))
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestDelete(t *testing.T) {
	repo := repository.NewMemoryRepository()
	created := repo.Create(newBook("a", "go"))

	if err := repo.Delete(created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.Get(created.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("Get after Delete: err = %v, want ErrNotFound", err)
	}
	if err := repo.Delete(created.ID); !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("second Delete: err = %v, want ErrNotFound", err)
	}
}

func TestList(t *testing.T) {
	repo := repository.NewMemoryRepository()
	// 25 books: odd ids are "go" (13 books), even ids are "py" (12 books).
	for i := 1; i <= 25; i++ {
		category := "py"
		if i%2 == 1 {
			category = "go"
		}
		repo.Create(newBook(fmt.Sprintf("book-%d", i), category))
	}

	tests := []struct {
		name     string
		category string
		page     int
		limit    int
		want     []uint32
	}{
		{"first page", "", 1, 10, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"second page", "", 2, 10, []uint32{11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
		{"last partial page", "", 3, 10, []uint32{21, 22, 23, 24, 25}},
		{"page past the data", "", 4, 10, []uint32{}},
		{"limit larger than data", "", 1, 100, nil}, // checked by length below
		{"filter first page", "go", 1, 5, []uint32{1, 3, 5, 7, 9}},
		{"filter second page", "go", 2, 5, []uint32{11, 13, 15, 17, 19}},
		{"filter last partial page", "go", 3, 5, []uint32{21, 23, 25}},
		{"filter without matches", "rust", 1, 10, []uint32{}},
		{"filter is case sensitive", "GO", 1, 10, []uint32{}},
		{"huge page does not overflow", "", math.MaxInt, 10, []uint32{}},
		{"huge page and huge limit", "", math.MaxInt, math.MaxInt, []uint32{}},
		{"huge limit", "", 1, math.MaxInt, nil}, // checked by length below
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repo.List(tt.category, tt.page, tt.limit)
			if got == nil {
				t.Fatal("List returned nil, want a non-nil slice (it must marshal to [])")
			}

			if tt.want == nil {
				if len(got) != 25 {
					t.Errorf("got %d books, want 25", len(got))
				}
				return
			}
			if gotIDs := ids(got); !slices.Equal(gotIDs, tt.want) {
				t.Errorf("ids = %v, want %v", gotIDs, tt.want)
			}
		})
	}
}

func TestList_EmptyRepositoryReturnsEmptySlice(t *testing.T) {
	repo := repository.NewMemoryRepository()

	got := repo.List("", 1, 10)
	if got == nil || len(got) != 0 {
		t.Errorf("List = %#v, want empty non-nil slice", got)
	}
}

func TestList_IsSortedByIDAfterDeletes(t *testing.T) {
	repo := repository.NewMemoryRepository()
	for i := 0; i < 50; i++ {
		repo.Create(newBook(fmt.Sprintf("book-%d", i), "go"))
	}
	for _, id := range []uint32{3, 17, 18, 40} {
		if err := repo.Delete(id); err != nil {
			t.Fatalf("Delete(%d): %v", id, err)
		}
	}

	// Map iteration order is random, so repeat to make an unsorted result likely to show up.
	for run := 0; run < 20; run++ {
		got := ids(repo.List("", 1, 100))
		if len(got) != 46 {
			t.Fatalf("got %d books, want 46", len(got))
		}
		if !slices.IsSorted(got) {
			t.Fatalf("ids are not sorted: %v", got)
		}
	}
}

func TestRepositoryStoresCopies(t *testing.T) {
	repo := repository.NewMemoryRepository()

	in := newBook("a", "go")
	in.PublishedYear = ptr(2015)
	in.Description = ptr("desc")
	created := repo.Create(in)

	// Changing the argument after Create must not touch the stored book.
	*in.PublishedYear = 1
	*in.Description = "changed via input"

	// Changing a returned book must not touch the stored book either.
	*created.PublishedYear = 2
	*created.Description = "changed via result"

	got, err := repo.Get(created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if *got.PublishedYear != 2015 || *got.Description != "desc" {
		t.Errorf("stored book was modified from outside: year=%d description=%q",
			*got.PublishedYear, *got.Description)
	}

	// The same applies to books returned by Get and List.
	*got.PublishedYear = 3
	again, _ := repo.Get(created.ID)
	if *again.PublishedYear != 2015 {
		t.Errorf("Get returned a pointer into the store: year=%d", *again.PublishedYear)
	}
	listed := repo.List("", 1, 10)
	*listed[0].PublishedYear = 4
	again, _ = repo.Get(created.ID)
	if *again.PublishedYear != 2015 {
		t.Errorf("List returned a pointer into the store: year=%d", *again.PublishedYear)
	}
}

func TestConcurrentCreatesGetUniqueIDs(t *testing.T) {
	repo := repository.NewMemoryRepository()

	const workers = 50
	got := make([]uint32, workers)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got[i] = repo.Create(newBook(fmt.Sprintf("book-%d", i), "go")).ID
			// Readers run alongside writers; -race checks this for data races.
			repo.List("", 1, 10)
			_, _ = repo.Get(got[i])
		}(i)
	}
	wg.Wait()

	slices.Sort(got)
	for i, id := range got {
		if want := uint32(i + 1); id != want {
			t.Fatalf("ids = %v, want 1..%d without gaps or duplicates", got, workers)
		}
	}
}
