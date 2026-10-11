package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"homework/internal/model"
)

func TestBook_JSON(t *testing.T) {
	created := time.Date(2026, 10, 5, 12, 0, 0, 123456000, time.UTC)

	t.Run("unset optional fields are null and still present", func(t *testing.T) {
		b := model.Book{ID: 1, Title: "T", ISBN: "I", Author: "A", Category: "C", CreatedAt: created, UpdatedAt: created}

		raw, err := json.Marshal(b)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatal(err)
		}

		for _, key := range []string{"published_year", "description"} {
			v, ok := got[key]
			if !ok {
				t.Errorf("key %q is missing, it must be present as null", key)
			} else if v != nil {
				t.Errorf("%s = %v, want null", key, v)
			}
		}
	})

	t.Run("keys are snake_case and time is RFC 3339", func(t *testing.T) {
		year, desc := 2015, "d"
		b := model.Book{ID: 1, Title: "T", ISBN: "I", Author: "A", Category: "C",
			PublishedYear: &year, Description: &desc, CreatedAt: created, UpdatedAt: created}

		raw, err := json.Marshal(b)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatal(err)
		}

		for _, key := range []string{"id", "title", "isbn", "author", "category",
			"published_year", "description", "created_at", "updated_at"} {
			if _, ok := got[key]; !ok {
				t.Errorf("key %q is missing", key)
			}
		}
		if got["created_at"] != "2026-10-05T12:00:00.123456Z" {
			t.Errorf("created_at = %v, want 2026-10-05T12:00:00.123456Z", got["created_at"])
		}
	})
}
