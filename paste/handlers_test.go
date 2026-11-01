package paste

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func newTestServer(t *testing.T) (*httptest.Server, *Store) {
	t.Helper()
	store := NewStore()
	h := NewHandlers(store)
	rl := NewRateLimiter(rate.Every(30*time.Second), 1)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/paste", rl.Middleware(h.Create))
	mux.HandleFunc("GET /api/paste/{id}", h.Get)
	mux.HandleFunc("GET /api/paste/{id}/raw", h.GetRaw)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, store
}

var clientIP atomic.Int64

// post sends a create request from a unique client IP so tests don't trip the rate limiter.
func post(t *testing.T, srv *httptest.Server, body string) (*http.Response, map[string]any) {
	t.Helper()
	req, _ := http.NewRequest("POST", srv.URL+"/api/paste", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "10.0.0."+strconv.FormatInt(clientIP.Add(1), 10))
	return do(t, req)
}

func get(t *testing.T, srv *httptest.Server, path, password string) (*http.Response, map[string]any) {
	t.Helper()
	req, _ := http.NewRequest("GET", srv.URL+path, nil)
	if password != "" {
		req.Header.Set("X-Paste-Password", password)
	}
	return do(t, req)
}

func do(t *testing.T, req *http.Request) (*http.Response, map[string]any) {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	return resp, body
}

func TestCreateAndGet(t *testing.T) {
	srv, _ := newTestServer(t)

	resp, body := post(t, srv, `{"title":"hi","content":"package main","language":"go","expiry":5}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: got %d %v", resp.StatusCode, body)
	}
	id, _ := body["id"].(string)
	if len(id) != 8 || body["url"] != "/paste/"+id {
		t.Fatalf("create: unexpected body %v", body)
	}

	resp, body = get(t, srv, "/api/paste/"+id, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get: got %d %v", resp.StatusCode, body)
	}
	for k, want := range map[string]any{"id": id, "title": "hi", "content": "package main", "language": "go", "protected": false} {
		if body[k] != want {
			t.Errorf("get: %s = %v, want %v", k, body[k], want)
		}
	}
	created, err1 := time.Parse(time.RFC3339, body["created_at"].(string))
	expires, err2 := time.Parse(time.RFC3339, body["expires_at"].(string))
	if err1 != nil || err2 != nil || expires.Sub(created) != 5*time.Minute {
		t.Errorf("get: bad timestamps created_at=%v expires_at=%v", body["created_at"], body["expires_at"])
	}
	if _, ok := body["PasswordHash"]; ok {
		t.Error("get: password hash leaked")
	}
}

func TestLanguageDefaultsToPlaintext(t *testing.T) {
	srv, _ := newTestServer(t)
	_, body := post(t, srv, `{"content":"x","expiry":1}`)
	_, body = get(t, srv, "/api/paste/"+body["id"].(string), "")
	if body["language"] != "plaintext" {
		t.Errorf("language = %v, want plaintext", body["language"])
	}
}

func TestRaw(t *testing.T) {
	srv, _ := newTestServer(t)
	_, body := post(t, srv, `{"content":"<script>hi</script>","expiry":1}`)

	resp, err := http.Get(srv.URL + "/api/paste/" + body["id"].(string) + "/raw")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(raw) != "<script>hi</script>" {
		t.Fatalf("raw: got %d %q", resp.StatusCode, raw)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("raw: Content-Type = %q", ct)
	}
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("raw: missing nosniff")
	}
}

func TestPasswordProtected(t *testing.T) {
	srv, _ := newTestServer(t)
	_, body := post(t, srv, `{"content":"secret","expiry":1,"password":"hunter2"}`)
	path := "/api/paste/" + body["id"].(string)

	for _, pw := range []string{"", "wrong"} {
		resp, body := get(t, srv, path, pw)
		if resp.StatusCode != http.StatusForbidden || body["protected"] != true || body["error"] != "password required" {
			t.Errorf("password %q: got %d %v", pw, resp.StatusCode, body)
		}
		resp, _ = get(t, srv, path+"/raw", pw)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("raw with password %q: got %d", pw, resp.StatusCode)
		}
	}

	resp, body := get(t, srv, path, "hunter2")
	if resp.StatusCode != http.StatusOK || body["content"] != "secret" || body["protected"] != true {
		t.Errorf("correct password: got %d %v", resp.StatusCode, body)
	}
}

func TestNotFoundAndExpired(t *testing.T) {
	srv, store := newTestServer(t)

	if resp, body := get(t, srv, "/api/paste/nope", ""); resp.StatusCode != http.StatusNotFound || body["error"] != "paste not found" {
		t.Errorf("missing: got %d %v", resp.StatusCode, body)
	}

	_, body := post(t, srv, `{"content":"x","expiry":1}`)
	id := body["id"].(string)
	p, _ := store.Get(id)
	p.ExpiresAt = time.Now().Add(-time.Second)

	if resp, body := get(t, srv, "/api/paste/"+id, ""); resp.StatusCode != http.StatusGone || body["error"] != "paste has expired" {
		t.Errorf("expired: got %d %v", resp.StatusCode, body)
	}
	if _, ok := store.Get(id); ok {
		t.Error("expired paste was not deleted on read")
	}
}

func TestDeleteExpired(t *testing.T) {
	store := NewStore()
	now := time.Now()
	old := &Paste{ExpiresAt: now.Add(-time.Second)}
	fresh := &Paste{ExpiresAt: now.Add(time.Minute)}
	store.Create(old)
	store.Create(fresh)

	if n := store.DeleteExpired(now); n != 1 {
		t.Errorf("DeleteExpired = %d, want 1", n)
	}
	if _, ok := store.Get(old.ID); ok {
		t.Error("expired paste still present")
	}
	if _, ok := store.Get(fresh.ID); !ok {
		t.Error("fresh paste was deleted")
	}
}

func TestCreateValidation(t *testing.T) {
	srv, _ := newTestServer(t)
	tests := []struct {
		name, body string
		status     int
		error      string
	}{
		{"bad json", `{`, 400, "invalid request body"},
		{"empty content", `{"content":"","expiry":5}`, 400, "content is required"},
		{"expiry too low", `{"content":"x","expiry":0}`, 400, "expiry must be between 1 and 10 minutes"},
		{"expiry too high", `{"content":"x","expiry":11}`, 400, "expiry must be between 1 and 10 minutes"},
		{"long title", `{"content":"x","expiry":5,"title":"` + strings.Repeat("a", 201) + `"}`, 400, "title exceeds 200 characters"},
		{"long password", `{"content":"x","expiry":5,"password":"` + strings.Repeat("a", 73) + `"}`, 400, "password exceeds 72 bytes"},
		{"content over 1 MB", `{"content":"` + strings.Repeat("a", maxContentBytes+1) + `","expiry":5}`, 413, "content exceeds 1 MB"},
		{"body over limit", `{"content":"` + strings.Repeat("a", maxBodyBytes) + `","expiry":5}`, 413, "content exceeds 1 MB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := post(t, srv, tt.body)
			if resp.StatusCode != tt.status || body["error"] != tt.error {
				t.Errorf("got %d %v, want %d %q", resp.StatusCode, body, tt.status, tt.error)
			}
		})
	}
}

func TestCreateLimitsAllowEdgeCases(t *testing.T) {
	srv, _ := newTestServer(t)
	tests := []struct{ name, body string }{
		// 200 multibyte characters is 800 bytes but still within the title limit.
		{"200 emoji title", `{"content":"x","expiry":5,"title":"` + strings.Repeat("🙂", 200) + `"}`},
		// Escaped newlines double the wire size; a full-size paste must still fit.
		{"1 MB of newlines", `{"content":"` + strings.Repeat(`\n`, maxContentBytes) + `","expiry":5}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if resp, body := post(t, srv, tt.body); resp.StatusCode != http.StatusCreated {
				t.Errorf("got %d %v", resp.StatusCode, body)
			}
		})
	}
}

func TestRateLimit(t *testing.T) {
	srv, _ := newTestServer(t)
	send := func() *http.Response {
		req, _ := http.NewRequest("POST", srv.URL+"/api/paste", strings.NewReader(`{"content":"x","expiry":1}`))
		req.Header.Set("X-Forwarded-For", "192.0.2.1, 10.0.0.1")
		resp, _ := do(t, req)
		return resp
	}
	if resp := send(); resp.StatusCode != http.StatusCreated {
		t.Fatalf("first: got %d", resp.StatusCode)
	}
	if resp := send(); resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("second: got %d, want 429", resp.StatusCode)
	}
}
