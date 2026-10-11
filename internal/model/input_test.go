package model_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"homework/internal/model"
)

func TestBookInput_DecodeAndValidate(t *testing.T) {
	const valid = `"title":"T","isbn":"I","author":"A","category":"C"`

	tests := []struct {
		name string
		body string
		// wantDecodeErr: json.Unmarshal itself must fail (wrong type, broken JSON).
		wantDecodeErr bool
		// wantInvalid lists fields that Validate must report; empty means valid.
		wantInvalid []string
	}{
		{name: "all required fields", body: `{` + valid + `}`},
		{name: "with optional fields", body: `{` + valid + `,"published_year":2015,"description":"d"}`},
		{name: "optional fields are null", body: `{` + valid + `,"published_year":null,"description":null}`},

		{name: "missing title", body: `{"isbn":"I","author":"A","category":"C"}`, wantInvalid: []string{"title"}},
		{name: "null title", body: `{"title":null,"isbn":"I","author":"A","category":"C"}`, wantInvalid: []string{"title"}},
		{name: "empty title", body: `{"title":"","isbn":"I","author":"A","category":"C"}`, wantInvalid: []string{"title"}},
		{name: "blank title", body: `{"title":"   \t","isbn":"I","author":"A","category":"C"}`, wantInvalid: []string{"title"}},
		{name: "missing isbn", body: `{"title":"T","author":"A","category":"C"}`, wantInvalid: []string{"isbn"}},
		{name: "missing author", body: `{"title":"T","isbn":"I","category":"C"}`, wantInvalid: []string{"author"}},
		{name: "missing category", body: `{"title":"T","isbn":"I","author":"A"}`, wantInvalid: []string{"category"}},
		{name: "empty object", body: `{}`, wantInvalid: []string{"title", "isbn", "author", "category"}},

		{name: "title is a number", body: `{"title":1,"isbn":"I","author":"A","category":"C"}`, wantDecodeErr: true},
		{name: "published_year is a string", body: `{` + valid + `,"published_year":"2015"}`, wantDecodeErr: true},
		{name: "published_year is fractional", body: `{` + valid + `,"published_year":1.5}`, wantDecodeErr: true},
		{name: "description is a number", body: `{` + valid + `,"description":5}`, wantDecodeErr: true},
		{name: "broken json", body: `{`, wantDecodeErr: true},
		{name: "empty body", body: ``, wantDecodeErr: true},
		{name: "json array", body: `[]`, wantDecodeErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var in model.BookInput
			err := json.Unmarshal([]byte(tt.body), &in)
			if (err != nil) != tt.wantDecodeErr {
				t.Fatalf("decode error = %v, wantDecodeErr %v", err, tt.wantDecodeErr)
			}
			if tt.wantDecodeErr {
				return
			}

			err = in.Validate()
			if len(tt.wantInvalid) == 0 {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, model.ErrValidation) {
				t.Fatalf("Validate() = %v, want an error wrapping ErrValidation", err)
			}
			for _, field := range tt.wantInvalid {
				if !strings.Contains(err.Error(), field) {
					t.Errorf("error %q does not mention %q", err, field)
				}
			}
		})
	}
}
