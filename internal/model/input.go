package model

import (
	"errors"
	"fmt"
	"strings"
)

// ErrValidation marks every validation failure. Callers detect it with
// errors.Is(err, ErrValidation) and answer 422 validation_error.
var ErrValidation = errors.New("validation error")

// BookInput is the request body for POST and PUT.
//
// Required fields are pointers so that a missing field or an explicit null
// (nil) can be told apart from an empty string. A value of the wrong JSON
// type is rejected by encoding/json during decoding.
type BookInput struct {
	Title         *string `json:"title"`
	ISBN          *string `json:"isbn"`
	Author        *string `json:"author"`
	Category      *string `json:"category"`
	PublishedYear *int    `json:"published_year"`
	Description   *string `json:"description"`
}

// Validate reports which required fields are missing, null, empty or blank.
// The returned error wraps ErrValidation.
func (in BookInput) Validate() error {
	required := []struct {
		name  string
		value *string
	}{
		{"title", in.Title},
		{"isbn", in.ISBN},
		{"author", in.Author},
		{"category", in.Category},
	}

	var invalid []string
	for _, f := range required {
		if f.value == nil || strings.TrimSpace(*f.value) == "" {
			invalid = append(invalid, f.name)
		}
	}
	if len(invalid) > 0 {
		return fmt.Errorf("%w: required fields are missing or empty: %s",
			ErrValidation, strings.Join(invalid, ", "))
	}
	return nil
}
