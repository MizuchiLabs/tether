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

func (s *Server) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Token == "" {
			next.ServeHTTP(w, r) // Authentication disabled
			return
		}

		token := util.GetAccessToken(r.Header)
		if token != "" && subtle.ConstantTimeCompare([]byte(s.cfg.Token), []byte(token)) == 1 {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

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

		// Verify the secret against your config
		if token != "" && subtle.ConstantTimeCompare([]byte(req.Secret), []byte(token)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Set the secure, HttpOnly cookie
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

func Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
