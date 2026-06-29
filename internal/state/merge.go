// Package state contains the traefik configuration by environment
package state

import (
	"log/slog"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// mergeMap copies entries from src into dst, skipping key collisions.
func mergeMap[K comparable, V any](dst, src map[K]V) map[K]V {
	if len(src) == 0 {
		return dst
	}

	if dst == nil {
		dst = make(map[K]V)
	}

	for k, v := range src {
		if _, exists := dst[k]; exists {
			slog.Warn("Collision detected, skipping", "key", k)
			continue
		}
		dst[k] = v
	}

	return dst
}

// mergeHTTP merges HTTP routers, middlewares, services, and transports.
func mergeHTTP(dst, src *dynamic.HTTPConfiguration) *dynamic.HTTPConfiguration {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = &dynamic.HTTPConfiguration{}
	}

	dst.Routers = mergeMap(dst.Routers, src.Routers)
	dst.Middlewares = mergeMap(dst.Middlewares, src.Middlewares)
	dst.Services = mergeMap(dst.Services, src.Services)
	dst.ServersTransports = mergeMap(dst.ServersTransports, src.ServersTransports)

	if dst.Routers == nil && dst.Middlewares == nil &&
		dst.Services == nil && dst.ServersTransports == nil {
		return nil
	}

	return dst
}

func mergeTCP(dst, src *dynamic.TCPConfiguration) *dynamic.TCPConfiguration {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = &dynamic.TCPConfiguration{}
	}

	dst.Routers = mergeMap(dst.Routers, src.Routers)
	dst.Middlewares = mergeMap(dst.Middlewares, src.Middlewares)
	dst.Services = mergeMap(dst.Services, src.Services)
	dst.ServersTransports = mergeMap(dst.ServersTransports, src.ServersTransports)

	if dst.Routers == nil && dst.Middlewares == nil &&
		dst.Services == nil && dst.ServersTransports == nil {
		return nil
	}

	return dst
}

func mergeUDP(dst, src *dynamic.UDPConfiguration) *dynamic.UDPConfiguration {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = &dynamic.UDPConfiguration{}
	}

	dst.Routers = mergeMap(dst.Routers, src.Routers)
	dst.Services = mergeMap(dst.Services, src.Services)

	if dst.Routers == nil && dst.Services == nil {
		return nil
	}

	return dst
}

func mergeTLS(dst, src *dynamic.TLSConfiguration) *dynamic.TLSConfiguration {
	if src == nil {
		return dst
	}
	if dst == nil {
		dst = &dynamic.TLSConfiguration{}
	}

	if len(src.Certificates) > 0 {
		dst.Certificates = append(dst.Certificates, src.Certificates...)
	}
	dst.Options = mergeMap(dst.Options, src.Options)
	dst.Stores = mergeMap(dst.Stores, src.Stores)

	if len(dst.Certificates) == 0 && dst.Options == nil && dst.Stores == nil {
		return nil
	}

	return dst
}
