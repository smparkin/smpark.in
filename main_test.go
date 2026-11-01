package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func fetch(t *testing.T, srv *httptest.Server, method, path string, header http.Header) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest(method, srv.URL+path, nil)
	for k, v := range header {
		req.Header[k] = v
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp, string(body)
}

func TestRoutes(t *testing.T) {
	srv := httptest.NewServer(newMux())
	defer srv.Close()
	year := strconv.Itoa(time.Now().Year())

	tests := []struct {
		method, path string
		status       int
		contentType  string
		contains     string
	}{
		{"GET", "/", 200, "text/html; charset=utf-8", "&copy; " + year + " Stephen Parkinson"},
		{"GET", "/index.html", 200, "text/html; charset=utf-8", "<h1>Stephen Parkinson</h1>"},
		{"GET", "/resume", 200, "text/html; charset=utf-8", "&copy; " + year + " Stephen Parkinson"},
		{"GET", "/css/style.css", 200, "text/css; charset=utf-8", "#9370db"},
		{"GET", "/images/me.jpg", 200, "image/jpeg", ""},
		{"GET", "/.well-known/appspecific/com.tesla.3p.public-key.pem", 200, "", "BEGIN PUBLIC KEY"},
		{"HEAD", "/", 200, "text/html; charset=utf-8", ""},
		{"GET", "/nope", 404, "text/html; charset=utf-8", "404 — Page Not Found"},
		{"GET", "/css/", 404, "text/html; charset=utf-8", "404 — Page Not Found"},
		{"GET", "/.DS_Store", 404, "text/html; charset=utf-8", "404 — Page Not Found"},
		{"GET", "/paste", 404, "text/html; charset=utf-8", "404 — Page Not Found"},
		{"GET", "/api/paste/AbCd1234", 404, "text/html; charset=utf-8", "404 — Page Not Found"},
		{"POST", "/resume", 405, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			resp, body := fetch(t, srv, tt.method, tt.path, nil)
			if resp.StatusCode != tt.status {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.status)
			}
			if tt.contentType != "" && resp.Header.Get("Content-Type") != tt.contentType {
				t.Errorf("Content-Type = %q, want %q", resp.Header.Get("Content-Type"), tt.contentType)
			}
			if !strings.Contains(body, tt.contains) {
				t.Errorf("body does not contain %q:\n%.300s", tt.contains, body)
			}
		})
	}
}

func TestStaticETag(t *testing.T) {
	srv := httptest.NewServer(newMux())
	defer srv.Close()

	resp, _ := fetch(t, srv, "GET", "/css/style.css", nil)
	etag := resp.Header.Get("ETag")
	if etag == "" {
		t.Fatal("no ETag on static file")
	}
	resp, _ = fetch(t, srv, "GET", "/css/style.css", http.Header{"If-None-Match": {etag}})
	if resp.StatusCode != http.StatusNotModified {
		t.Errorf("revalidation: got %d, want 304", resp.StatusCode)
	}
}

func TestLogRequests(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	h := logRequests(newMux())

	viaTunnel := httptest.NewRequest("GET", "/nope", nil)
	viaTunnel.Header.Set("CF-Connecting-IP", "198.51.100.7")
	h.ServeHTTP(httptest.NewRecorder(), viaTunnel)

	direct := httptest.NewRequest("GET", "/resume", nil)
	direct.RemoteAddr = "192.168.1.50:51234"
	h.ServeHTTP(httptest.NewRecorder(), direct)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d log lines, want 2:\n%s", len(lines), buf.String())
	}
	for i, want := range [][]string{
		{"msg=request", "method=GET", "path=/nope", "status=404", "client=198.51.100.7", "duration="},
		{"msg=request", "method=GET", "path=/resume", "status=200", "client=192.168.1.50:51234"},
	} {
		for _, field := range want {
			if !strings.Contains(lines[i], field) {
				t.Errorf("line %d missing %q: %s", i, field, lines[i])
			}
		}
	}
}
