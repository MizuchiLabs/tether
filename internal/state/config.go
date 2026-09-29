package state

// Config is a traefik dynamic configuration as plain JSON values. Tether only reads the
// layout (sections, entry maps, router services, load balancer servers), everything else
// passes through untouched, so it works with any traefik version.
type Config map[string]any

// section returns a top level section like "http", or nil.
func (c Config) section(name string) map[string]any {
	return asMap(c[name])
}

// count returns the number of routers and services across all protocols.
func (c Config) count() (routers, services int) {
	for _, proto := range []string{"http", "tcp", "udp"} {
		s := c.section(proto)
		routers += len(asMap(s["routers"]))
		services += len(asMap(s["services"]))
	}
	return routers, services
}

func asMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}
