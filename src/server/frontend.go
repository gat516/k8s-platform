package server

import (
	"io/fs"
	"net/http"
	"os"
	"strings"
)

// dashboardFiles exposes the one-page UI and its compiled assets. Unknown API
// paths remain 404s; directory listings and arbitrary files are never served.
func dashboardFiles(dir string) http.Handler {
	files := os.DirFS(dir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "index.html"
		} else if !strings.HasPrefix(name, "assets/") || !fs.ValidPath(name) {
			http.NotFound(w, r)
			return
		}
		info, err := fs.Stat(files, name)
		if err != nil || !info.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}
		if name == "index.html" {
			w.Header().Set("Cache-Control", "no-cache")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeFileFS(w, r, files, name)
	})
}
