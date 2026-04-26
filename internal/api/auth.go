package api

import (
	"errors"
	"net/http"

	"github.com/mizuchilabs/tether/internal/config"
	"github.com/mizuchilabs/tether/internal/util"
)

var ErrUnauthorized = errors.New("unauthorized")

type AuthInterceptor struct {
	cfg *config.Config
}

func NewAuthInterceptor(cfg *config.Config) *AuthInterceptor {
	return &AuthInterceptor{cfg: cfg}
}

func (a *AuthInterceptor) WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := a.authenticate(r.Header); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authenticate routes to the appropriate authentication method
func (a *AuthInterceptor) authenticate(header http.Header) error {
	if a.cfg.Secret == "" {
		return nil // Authentication disabled
	}

	bearer := util.GetBearerToken(header)
	if bearer == "" {
		return ErrUnauthorized
	}
	if a.cfg.Secret != bearer {
		return ErrUnauthorized
	}
	return nil
}
