package auth

import (
	"net/http"
	"testing"
)

func TestGetAPIKey_Success(t *testing.T) {
	h := http.Header{}
	h.Set("Authorization", "ApiKey abc123")

	key, err := GetAPIKey(h)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if key != "abc123" {
		t.Fatalf("expected key 'abc123', got %q", key)
	}
}

func TestGetAPIKey_MissingHeader(t *testing.T) {
	h := http.Header{} // kein Authorization-Header

	if key, err := GetAPIKey(h); err == nil {
		t.Fatalf("expected error, got key=%q", key)
	}
}

func TestGetAPIKey_WrongScheme(t *testing.T) {
	h := http.Header{}
	h.Set("Authorization", "Bearer sometoken")

	if key, err := GetAPIKey(h); err == nil {
		t.Fatalf("expected error, got key=%q", key)
	}
}
