package api

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

const accessCookie = "tether_access"

// withAuth checks the bearer token or access cookie before calling the next handler.
func withAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" || validToken(token, requestToken(r)) {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

// login validates the secret and sets an access cookie.
func login(token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Secret string `json:"secret"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if token != "" && !validToken(token, req.Secret) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		setAccessCookie(w, r, req.Secret, 86400*7)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	setAccessCookie(w, r, "", -1)
}

// setAccessCookie marks the cookie Secure only over TLS, browsers drop Secure cookies on plain http.
func setAccessCookie(w http.ResponseWriter, r *http.Request, value string, maxAge int) {
	// #nosec G124 -- Secure is set whenever the request came over TLS
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookie,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	})
}

func requestToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	// Scheme is case insensitive, see RFC 9110 Section 11.1.
	if scheme, token, ok := strings.Cut(auth, " "); ok && strings.EqualFold(scheme, "Bearer") {
		return token
	}
	if c, err := r.Cookie(accessCookie); err == nil {
		return c.Value
	}
	return ""
}

func validToken(expected, got string) bool {
	return got != "" && subtle.ConstantTimeCompare([]byte(expected), []byte(got)) == 1
}
