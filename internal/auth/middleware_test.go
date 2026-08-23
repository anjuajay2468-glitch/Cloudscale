package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareAllowsAdminGET(t *testing.T) {
	authenticator := NewWithKeys(map[string]APIKey{
		"admin": {
			Key:  []byte("admin-key"),
			Role: RoleAdmin,
		},
	})

	handler := authenticator.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/objects/test.txt",
		nil,
	)

	req.Header.Set(APIKeyHeader, "admin-key")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddlewareAllowsReaderGET(t *testing.T) {
	authenticator := NewWithKeys(map[string]APIKey{
		"reader": {
			Key:  []byte("reader-key"),
			Role: RoleReader,
		},
	})

	handler := authenticator.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/objects/test.txt",
		nil,
	)

	req.Header.Set(APIKeyHeader, "reader-key")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMiddlewareRejectsReaderPUT(t *testing.T) {
	authenticator := NewWithKeys(map[string]APIKey{
		"reader": {
			Key:  []byte("reader-key"),
			Role: RoleReader,
		},
	})

	handler := authenticator.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("protected handler should not be called")
		}),
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/objects/test.txt",
		nil,
	)

	req.Header.Set(APIKeyHeader, "reader-key")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestMiddlewareRejectsReaderDELETE(t *testing.T) {
	authenticator := NewWithKeys(map[string]APIKey{
		"reader": {
			Key:  []byte("reader-key"),
			Role: RoleReader,
		},
	})

	handler := authenticator.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("protected handler should not be called")
		}),
	)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/objects/test.txt",
		nil,
	)

	req.Header.Set(APIKeyHeader, "reader-key")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestMiddlewareRejectsMissingKey(t *testing.T) {
	authenticator := NewWithKeys(map[string]APIKey{
		"admin": {
			Key:  []byte("admin-key"),
			Role: RoleAdmin,
		},
	})

	handler := authenticator.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("protected handler should not be called")
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/objects/test.txt",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMiddlewareRejectsWrongKey(t *testing.T) {
	authenticator := NewWithKeys(map[string]APIKey{
		"admin": {
			Key:  []byte("admin-key"),
			Role: RoleAdmin,
		},
	})

	handler := authenticator.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("protected handler should not be called")
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/objects/test.txt",
		nil,
	)

	req.Header.Set(APIKeyHeader, "wrong-key")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
