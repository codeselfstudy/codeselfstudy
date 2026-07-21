package digest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPPosterSuccess(t *testing.T) {
	var gotBody []byte
	var gotContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	if err := NewHTTPPoster(srv.URL).Post(context.Background(), []byte(`{"x":1}`)); err != nil {
		t.Fatalf("Post: %v", err)
	}
	if string(gotBody) != `{"x":1}` {
		t.Errorf("body = %s", gotBody)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q", gotContentType)
	}
}

func TestHTTPPosterNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()

	err := NewHTTPPoster(srv.URL).Post(context.Background(), []byte(`{}`))
	if err == nil {
		t.Fatal("expected error on 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %v, want to mention 500", err)
	}
}

func TestHTTPPosterNetworkError(t *testing.T) {
	// Nothing listens on port 1; Post must return an error, not hang. The webhook
	// URL is a secret (a bearer credential), so the error must NOT leak it.
	const secretURL = "http://127.0.0.1:1/services/T00/B00/SUPERSECRETTOKEN"
	err := NewHTTPPoster(secretURL).Post(context.Background(), []byte(`{}`))
	if err == nil {
		t.Fatal("expected network error")
	}
	if strings.Contains(err.Error(), "SUPERSECRETTOKEN") || strings.Contains(err.Error(), "/services/") {
		t.Errorf("error leaks the webhook URL: %v", err)
	}
}
