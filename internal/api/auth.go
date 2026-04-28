package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mizuchilabs/tether/internal/util"
)

var ErrUnauthorized = errors.New("unauthorized")

type AuthService struct {
	secret string
}

type LoginRequest struct {
	Secret string `json:"secret"`
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{secret: secret}
}

func (a *AuthService) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := a.authenticate(r.Header); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authenticate routes to the appropriate authentication method
func (a *AuthService) authenticate(header http.Header) error {
	if a.secret == "" {
		return nil // Authentication disabled
	}

	token := util.GetAccessToken(header)
	if token == "" {
		return ErrUnauthorized
	}
	if a.secret != token {
		return ErrUnauthorized
	}
	return nil
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
		if token != "" && req.Secret != token {
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
