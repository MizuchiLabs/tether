// Package util contains utility functions
package util

import (
	"net/http"
	"strings"
)

const AccessTokenName = "tether_access"

func GetAccessToken(header http.Header) string {
	if token := getBearer(header); token != "" {
		return token
	}
	return getCookie(header, AccessTokenName)
}

func getBearer(header http.Header) string {
	const prefix = "Bearer "
	auth := header.Get("Authorization")
	// Case insensitive prefix match. See RFC 9110 Section 11.1.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return ""
	}
	return auth[len(prefix):]
}

func getCookie(header http.Header, name string) string {
	cookieHeader := header.Get("Cookie")
	if cookieHeader == "" {
		return ""
	}
	cookies, err := http.ParseCookie(cookieHeader)
	if err != nil {
		return ""
	}
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}
