package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"homework/internal/docs"
	"homework/internal/handler"
)

func TestDocsEndpoints(t *testing.T) {
	r := gin.New()
	r.GET("/openapi.yaml", handler.OpenAPI)
	r.GET("/swagger", handler.SwaggerUI)

	tests := []struct {
		name        string
		path        string
		contentType string
		check       func(body string) bool
	}{
		{"spec", "/openapi.yaml", "application/yaml", func(b string) bool { return b == string(docs.Spec) }},
		{"swagger ui", "/swagger", "text/html", func(b string) bool { return strings.Contains(b, "/openapi.yaml") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, tt.contentType) {
				t.Errorf("Content-Type = %q, want prefix %q", ct, tt.contentType)
			}
			if !tt.check(rec.Body.String()) {
				t.Errorf("unexpected body: %.80s...", rec.Body)
			}
		})
	}
}
