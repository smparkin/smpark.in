package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	publicDir := "./public"
	spaIndex := filepath.Join(publicDir, "app", "index.html")

	fs := http.FileServer(http.Dir(publicDir))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the static file
		path := filepath.Join(publicDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}

		// Fall back to SPA index for client-side routing
		http.ServeFile(w, r, spaIndex)
	})

	addr := ":8080"
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
