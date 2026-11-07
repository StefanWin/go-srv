package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestParseHeaders(t *testing.T) {
	got := parseHeaders(" X-Frame-Options : DENY ,Cache-Control:max-age=60,broken,X-Url:http://a")
	want := map[string]string{
		"X-Frame-Options": "DENY",
		"Cache-Control":   "max-age=60",
		"X-Url":           "http://a",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %q, want %q", k, got[k], v)
		}
	}
}

func TestBrowserURL(t *testing.T) {
	cases := map[string]string{
		"localhost": "http://localhost:8080",
		"0.0.0.0":   "http://localhost:8080",
		"::":        "http://localhost:8080",
		"":          "http://localhost:8080",
		"127.0.0.1": "http://127.0.0.1:8080",
		"::1":       "http://[::1]:8080",
	}
	for bind, want := range cases {
		if got := browserURL(bind, 8080); got != want {
			t.Errorf("browserURL(%q) = %q, want %q", bind, got, want)
		}
	}
}

func TestCustomHeadersOverrideDefaults(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := buildHandler(dir, map[string]string{
		"Cache-Control":               "max-age=60",
		"Access-Control-Allow-Origin": "https://example.com",
	}, true, false)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/a.txt", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "max-age=60" {
		t.Errorf("Cache-Control = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	h := buildHandler(t.TempDir(), nil, true, false)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("missing Access-Control-Allow-Origin")
	}
}
