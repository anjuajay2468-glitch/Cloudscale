package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInternalMiddlewareAcceptsCorrectToken(t *testing.T) {
	handler := InternalMiddleware(
		"internal-secret",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/internal/test",
		nil,
	)

	req.Header.Set(
		InternalTokenHeader,
		"internal-secret",
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestInternalMiddlewareRejectsMissingToken(t *testing.T) {
	handler := InternalMiddleware(
		"internal-secret",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("protected handler should not be called")
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/internal/test",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestInternalMiddlewareRejectsWrongToken(t *testing.T) {
	handler := InternalMiddleware(
		"internal-secret",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("protected handler should not be called")
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/internal/test",
		nil,
	)

	req.Header.Set(
		InternalTokenHeader,
		"wrong-secret",
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestInternalMiddlewareRejectsEmptyConfiguration(t *testing.T) {
	handler := InternalMiddleware(
		"",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("protected handler should not be called")
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/internal/test",
		nil,
	)

	req.Header.Set(
		InternalTokenHeader,
		"anything",
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
