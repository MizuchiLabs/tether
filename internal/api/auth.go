package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
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
		if a.secret == "" {
			next.ServeHTTP(w, r) // Authentication disabled
			return
		}

		// Try standard token/cookie auth first (for Web UI & browsers)
		token := util.GetAccessToken(r.Header)
		if token != "" && subtle.ConstantTimeCompare([]byte(a.secret), []byte(token)) == 1 {
			next.ServeHTTP(w, r)
			return
		}

		// Fall back to Agent HMAC authentication
		signatureHex := r.Header.Get("X-Signature")
		if signatureHex == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Read the body to calculate the hash
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		mac := hmac.New(sha256.New, []byte(a.secret))
		mac.Write(bodyBytes)
		expectedMAC := mac.Sum(nil)
		providedMAC, err := hex.DecodeString(signatureHex)
		if err != nil {
			http.Error(w, "Invalid signature format", http.StatusBadRequest)
			return
		}

		// Compare the signatures
		if !hmac.Equal(providedMAC, expectedMAC) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
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
