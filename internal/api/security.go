package api

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
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

// clientIPKey extracts and canonicalizes the IP resolved by Chi.
// httprate.CanonicalizeIP groups IPv6 by /64 to prevent SLAAC rotation bypasses.
func clientIPKey(r *http.Request) (string, error) {
	ip := middleware.GetClientIP(r.Context())
	return httprate.CanonicalizeIP(ip), nil
}

// rateLimitAPI limits /api requests per client IP.
func rateLimitAPI(n int, window time.Duration) func(http.Handler) http.Handler {
	clientIP := clientIPMiddleware()
	return func(next http.Handler) http.Handler {
		limited := clientIP(httprate.LimitBy(n, window, clientIPKey)(next))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}
			limited.ServeHTTP(w, r)
		})
	}
}

func clientIPMiddleware() func(http.Handler) http.Handler {
	spec := strings.TrimSpace(os.Getenv("TETHER_TRUSTED_PROXIES"))
	switch strings.ToLower(spec) {
	case "", "direct", "none":
		return middleware.ClientIPFromRemoteAddr
	case "cloudflare":
		return middleware.ClientIPFromHeader("CF-Connecting-IP")
	case "traefik", "nginx", "standard":
		return middleware.ClientIPFromHeader("X-Real-IP")

	default:
		// Comma-separated CIDR ranges or proxy hop count
		cidrs := strings.Split(spec, ",")
		for i := range cidrs {
			cidrs[i] = strings.TrimSpace(cidrs[i])
		}
		return middleware.ClientIPFromXFF(cidrs...)
	}
}
