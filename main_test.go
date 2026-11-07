package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
	}, true, false, false)

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
	h := buildHandler(t.TempDir(), nil, true, false, false)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("missing Access-Control-Allow-Origin")
	}
}

func TestDotfiles(t *testing.T) {
	dir := t.TempDir()
	for _, p := range []string{".env", ".git/config", "sub/.secret", "sub/ok.txt", "visible.txt"} {
		full := filepath.Join(dir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(p), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	get := func(h http.Handler, path string) (int, string) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		body, _ := io.ReadAll(rec.Body)
		return rec.Code, string(body)
	}

	hidden := buildHandler(dir, nil, false, false, false)
	for _, p := range []string{"/.env", "/.git/config", "/.git/", "/sub/.secret"} {
		if code, _ := get(hidden, p); code != http.StatusNotFound {
			t.Errorf("hidden: GET %s = %d, want 404", p, code)
		}
	}
	for _, p := range []string{"/visible.txt", "/sub/ok.txt"} {
		if code, _ := get(hidden, p); code != http.StatusOK {
			t.Errorf("hidden: GET %s = %d, want 200", p, code)
		}
	}
	for path, listing := range map[string][]string{"/": {".env", ".git"}, "/sub/": {".secret"}} {
		code, body := get(hidden, path)
		if code != http.StatusOK {
			t.Fatalf("hidden: GET %s = %d", path, code)
		}
		for _, name := range listing {
			if strings.Contains(body, name) {
				t.Errorf("hidden: listing of %s contains %s", path, name)
			}
		}
	}
	if _, body := get(hidden, "/"); !strings.Contains(body, "visible.txt") {
		t.Error("hidden: listing of / is missing visible.txt")
	}

	shown := buildHandler(dir, nil, false, false, true)
	for _, p := range []string{"/.env", "/.git/config", "/sub/.secret"} {
		if code, _ := get(shown, p); code != http.StatusOK {
			t.Errorf("dotfiles on: GET %s = %d, want 200", p, code)
		}
	}
	if _, body := get(shown, "/"); !strings.Contains(body, ".env") {
		t.Error("dotfiles on: listing of / is missing .env")
	}
}
