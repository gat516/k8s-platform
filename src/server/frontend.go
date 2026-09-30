package server

import (
	"bytes"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"
)

// Load the compiled UI once. Requests select a known asset; they never open a
// filesystem path. Symlinks, directories and files outside assets are excluded.
func dashboardFiles(dir string) http.Handler {
	type asset struct {
		body     []byte
		modified time.Time
	}
	assets := make(map[string]asset)
	files := os.DirFS(dir)
	_ = fs.WalkDir(files, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || !entry.Type().IsRegular() || (name != "index.html" && !strings.HasPrefix(name, "assets/")) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		body, err := fs.ReadFile(files, name)
		if err == nil {
			assets[name] = asset{body, info.ModTime()}
		}
		return nil
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "index.html"
		}
		file, ok := assets[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if name == "index.html" {
			w.Header().Set("Cache-Control", "no-cache")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeContent(w, r, name, file.modified, bytes.NewReader(file.body))
	})
}
