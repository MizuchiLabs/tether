package api

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/mizuchilabs/tether/internal/util"
)

type LoginRequest struct {
	Secret string `json:"secret"`
}

// WithAuth checks the request token before calling the next handler.
func WithAuth(expectedToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			token := util.GetAccessToken(r.Header)
			if token != "" && subtle.ConstantTimeCompare([]byte(expectedToken), []byte(token)) == 1 {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

// Login validates the secret and sets an access cookie.
func Login(token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if token != "" && subtle.ConstantTimeCompare([]byte(req.Secret), []byte(token)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     util.AccessTokenName,
			Value:    req.Secret,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   86400 * 7,
		})
		w.WriteHeader(http.StatusOK)
	}
}

// Logout clears the access cookie.
func Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     util.AccessTokenName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1,
		})
		w.WriteHeader(http.StatusOK)
	}
}
