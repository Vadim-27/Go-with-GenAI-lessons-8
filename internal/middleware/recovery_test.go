package middleware_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"homework/internal/middleware"
	"homework/internal/model"
)

const secret = "super-secret-internal-detail"

func newRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.GET("/panic", func(c *gin.Context) { panic(secret) })
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "fine") })
	r.GET("/panic-after-write", func(c *gin.Context) {
		c.String(http.StatusOK, "partial")
		panic(secret)
	})
	return r
}

func get(r http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// captureLog перенаправляє стандартний лог у буфер на час тесту.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &buf
}

func TestRecovery_PanicBecomes500WithoutDetails(t *testing.T) {
	logs := captureLog(t)

	rec := get(newRouter(), "/panic")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp model.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body is not the JSON error envelope: %v (%s)", err, rec.Body)
	}
	if resp.Error.Code != model.CodeInternalServerError {
		t.Errorf("code = %q, want %q", resp.Error.Code, model.CodeInternalServerError)
	}
	if resp.Error.Message == "" {
		t.Error("message must not be empty")
	}

	// Клієнт не має бачити ні текст паніки, ні stack trace.
	body := rec.Body.String()
	if strings.Contains(body, secret) || strings.Contains(body, "goroutine") {
		t.Errorf("response leaks internal details: %s", body)
	}

	// Подробиці мають потрапити в лог.
	if !strings.Contains(logs.String(), secret) {
		t.Errorf("panic details were not logged: %q", logs.String())
	}
}

func TestRecovery_ServerKeepsWorkingAfterPanic(t *testing.T) {
	captureLog(t)
	r := newRouter()

	if rec := get(r, "/panic"); rec.Code != http.StatusInternalServerError {
		t.Fatalf("panic request: status = %d, want 500", rec.Code)
	}

	rec := get(r, "/ok")
	if rec.Code != http.StatusOK || rec.Body.String() != "fine" {
		t.Errorf("next request: got %d %q, want 200 \"fine\"", rec.Code, rec.Body)
	}
}

func TestRecovery_PanicAfterResponseStarted(t *testing.T) {
	captureLog(t)

	rec := get(newRouter(), "/panic-after-write")

	// Статус і тіло вже відправлені, тому JSON-помилку дописувати не можна.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (already written)", rec.Code)
	}
	if rec.Body.String() != "partial" {
		t.Errorf("body = %q, want %q", rec.Body, "partial")
	}
}
