package model

import "time"

// Book is a stored book. Optional fields are pointers so that "not set" is
// serialised as JSON null instead of a zero value.
type Book struct {
	ID            uint32    `json:"id"`
	Title         string    `json:"title"`
	ISBN          string    `json:"isbn"`
	Author        string    `json:"author"`
	Category      string    `json:"category"`
	PublishedYear *int      `json:"published_year"`
	Description   *string   `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
