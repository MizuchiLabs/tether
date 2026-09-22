// Package web provides the embedded frontend build files.
package web

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

//go:generate pnpm install
//go:generate pnpm build
//go:embed all:build
var StaticFS embed.FS

func Handler() http.Handler {
	pinMimeTypes()
	files := statigz.FileServer(StaticFS, brotli.AddEncoding, statigz.FSPrefix("build"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p != "/" && !exists(p) {
			if exists(p + ".html") {
				p += ".html"
			} else {
				p = "/"
			}
		}
		r.URL.Path = p
		w.Header().Set("Cache-Control", cachePolicy(p))
		files.ServeHTTP(w, r)
	})
}

func pinMimeTypes() {
	_ = mime.AddExtensionType(".txt", "text/plain; charset=utf-8")
	_ = mime.AddExtensionType(".ico", "image/x-icon")
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
	_ = mime.AddExtensionType(".woff", "font/woff")
	_ = mime.AddExtensionType(".woff2", "font/woff2")
}

func exists(p string) bool {
	st, err := fs.Stat(StaticFS, path.Join("build", p))
	return err == nil && !st.IsDir()
}

func cachePolicy(p string) string {
	switch {
	case strings.HasPrefix(p, "/_app/immutable/"):
		return "public, max-age=31536000, immutable"
	case p == "/" || strings.HasSuffix(p, ".html") || p == "/service-worker.js" || p == "/_app/version.json":
		return "no-cache, must-revalidate"
	default:
		return "public, max-age=3600"
	}
}
