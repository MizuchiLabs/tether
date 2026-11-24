package api

import (
	"net/http"
	"strings"

	"github.com/unrolled/secure"
)

func securityHeaders() func(http.Handler) http.Handler {
	return secure.New(secure.Options{
		ContentTypeNosniff: true,
		FrameDeny:          true,
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		ContentSecurityPolicy: strings.Join([]string{
			"default-src 'self'",
			"base-uri 'self'",
			"object-src 'none'",
			"frame-ancestors 'none'",
			"form-action 'self'",
			"script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval'",
			"style-src 'self' 'unsafe-inline'",
			"img-src 'self' data: blob:",
			"font-src 'self' data:",
			"connect-src 'self'",
			"worker-src 'self' blob:",
		}, "; "),
	}).Handler
}
