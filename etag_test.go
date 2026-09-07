package etag_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	etag "github.com/pablor21/echo-etag/v5"
)

var testWeakEtag = "W/\"11-8dcfee46\""
var testStrEtag = "\"11-0a4d55a8d778e5022fab701977c5d840bbc486d0\""
var e *echo.Echo

func init() {
	e = echo.New()
	e.GET("/etag", func(c *echo.Context) error {
		return c.String(200, "Hello World")
	}, etag.WithConfig(etag.Config{Weak: false}))

	e.GET("/etag/weak", func(c *echo.Context) error {
		return c.String(200, "Hello World")
	}, etag.Etag())

	e.GET("/etag/nocontent", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}, etag.Etag())

	e.GET("/etag/unwrap", func(c *echo.Context) error {
		if _, err := echo.UnwrapResponse(c.Response()); err != nil {
			return err
		}
		return c.String(200, "Hello World")
	}, etag.Etag())

}

func TestStrongEtag(t *testing.T) {

	// Test strong Etag
	req := httptest.NewRequest(http.MethodGet, "/etag", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Etag") != testStrEtag {
		t.Errorf("Expected Etag %s, got %s", testStrEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "Hello World" {
		t.Errorf("Expected body %s, got %s", "Hello World", rec.Body.String())
	}

	// Test If-None-Match
	req = httptest.NewRequest(http.MethodGet, "/etag", nil)
	req.Header.Set("If-None-Match", testStrEtag)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified {
		t.Errorf("Expected status code %d, got %d", http.StatusNotModified, rec.Code)
	}

	if rec.Header().Get("Etag") != testStrEtag {
		t.Errorf("Expected Etag %s, got %s", testStrEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "" {
		t.Errorf("Expected body %s, got %s", "", rec.Body.String())
	}

	// Test If-None-Match invalid
	req = httptest.NewRequest(http.MethodGet, "/etag", nil)
	req.Header.Set("If-None-Match", "invalid")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Etag") != testStrEtag {
		t.Errorf("Expected Etag %s, got %s", testStrEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "Hello World" {
		t.Errorf("Expected body %s, got %s", "Hello World", rec.Body.String())
	}

}

func TestWeakEtag(t *testing.T) {

	// Test weak Etag
	req := httptest.NewRequest(http.MethodGet, "/etag/weak", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Etag") != testWeakEtag {
		t.Errorf("Expected Etag %s, got %s", testWeakEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "Hello World" {
		t.Errorf("Expected body %s, got %s", "Hello World", rec.Body.String())
	}

	// Test If-None-Match weak
	req = httptest.NewRequest(http.MethodGet, "/etag/weak", nil)
	req.Header.Set("If-None-Match", testWeakEtag)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotModified {
		t.Errorf("Expected status code %d, got %d", http.StatusNotModified, rec.Code)
	}

	if rec.Header().Get("Etag") != testWeakEtag {
		t.Errorf("Expected Etag %s, got %s", testWeakEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "" {
		t.Errorf("Expected body %s, got %s", "", rec.Body.String())
	}

	// Test If-None-Match weak invalid
	req = httptest.NewRequest(http.MethodGet, "/etag/weak", nil)
	req.Header.Set("If-None-Match", "invalid")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Etag") != testWeakEtag {
		t.Errorf("Expected Etag %s, got %s", testWeakEtag, rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "Hello World" {
		t.Errorf("Expected body %s, got %s", "Hello World", rec.Body.String())
	}

}

// TestNoContent ensures a handler that writes a status but no body passes through untouched and
// does not get an Etag.
func TestNoContent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/etag/nocontent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, rec.Code)
	}

	if rec.Header().Get("Etag") != "" {
		t.Errorf("Expected no Etag, got %s", rec.Header().Get("Etag"))
	}

	if rec.Body.String() != "" {
		t.Errorf("Expected body %s, got %s", "", rec.Body.String())
	}
}

// TestUnwrapResponse ensures the buffering writer installed by the middleware still unwraps to the
// context original *echo.Response, as `Context#SetResponse` requires.
func TestUnwrapResponse(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/etag/unwrap", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Etag") != testWeakEtag {
		t.Errorf("Expected Etag %s, got %s", testWeakEtag, rec.Header().Get("Etag"))
	}
}
