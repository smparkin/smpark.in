// Package site serves the HTML pages and static files, all embedded in the binary.
package site

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

// Dot-prefixed paths are skipped unless named, so .well-known is listed
// explicitly. That keeps strays like .DS_Store out of the binary.
//
//go:embed static static/.well-known
var staticFS embed.FS

var (
	templates  = template.Must(template.ParseFS(templateFS, "templates/*.html"))
	staticRoot = mustSub(staticFS, "static")
	fileServer = http.FileServerFS(staticRoot)

	// Embedded files have no modification time, so http.FileServer can't send
	// Last-Modified. Content-hash ETags let browsers revalidate instead.
	// The map also serves as the list of files that exist.
	etags = hashFiles(staticRoot)
)

// Register adds the page routes to mux, plus a catch-all that serves static
// files and renders the 404 page for everything else.
func Register(mux *http.ServeMux) {
	home := func(w http.ResponseWriter, r *http.Request) {
		render(w, http.StatusOK, "home.html", yearData())
	}
	mux.HandleFunc("GET /{$}", home)
	mux.HandleFunc("GET /index.html", home)
	mux.HandleFunc("GET /resume", func(w http.ResponseWriter, r *http.Request) {
		render(w, http.StatusOK, "resume.html", yearData())
	})
	mux.HandleFunc("GET /paste", func(w http.ResponseWriter, r *http.Request) {
		render(w, http.StatusOK, "paste.html", nil)
	})
	mux.HandleFunc("GET /paste/{id}", func(w http.ResponseWriter, r *http.Request) {
		render(w, http.StatusOK, "paste-view.html", struct{ PasteID string }{r.PathValue("id")})
	})
	mux.HandleFunc("GET /", serveStatic)
}

func yearData() any {
	return struct{ Year int }{time.Now().Year()}
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	etag, ok := etags[name]
	if !ok {
		render(w, http.StatusNotFound, "404.html", nil)
		return
	}
	w.Header().Set("ETag", etag)
	fileServer.ServeHTTP(w, r)
}

// render executes the template into a buffer first so a template error
// produces a clean 500 instead of a half-written page.
func render(w http.ResponseWriter, status int, name string, data any) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		slog.Error("render template", "name", name, "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}

func hashFiles(fsys fs.FS) map[string]string {
	etags := make(map[string]string)
	err := fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		etags[name] = `"` + hex.EncodeToString(sum[:8]) + `"`
		return nil
	})
	if err != nil {
		panic(err)
	}
	return etags
}
