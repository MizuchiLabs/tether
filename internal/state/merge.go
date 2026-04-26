package state

import (
	"log/slog"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// mergeMap handles lazy initialization and collision detection for ANY map type.
func mergeMap[K comparable, V any](dst, src map[K]V) map[K]V {
	// If source is empty, do nothing
	if len(src) == 0 {
		return dst
	}

	// Lazily initialize destination only if we actually have data to merge
	if dst == nil {
		dst = make(map[K]V)
	}

	// Merge and check for collisions
	for k, v := range src {
		if _, exists := dst[k]; exists {
			slog.Warn("Collision detected, skipping", "key", k)
			continue
		}
		dst[k] = v
	}

	return dst
}

// mergeHTTP merges HTTP configurations safely and lazily.
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
