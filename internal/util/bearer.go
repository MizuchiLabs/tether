// Package util contains utility functions
package util

import (
	"net/http"
	"strings"
)

func GetBearerToken(header http.Header) string {
	const prefix = "Bearer "
	auth := header.Get("Authorization")
	// Case insensitive prefix match. See RFC 9110 Section 11.1.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return ""
	}
	return auth[len(prefix):]
}
