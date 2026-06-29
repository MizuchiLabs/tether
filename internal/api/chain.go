package api

import (
	"net/http"
)

// Constructor wraps middleware that transforms an http.Handler.
type Constructor func(http.Handler) http.Handler

// Chain composes middleware constructors into a pipeline.
type Chain struct {
	constructors []Constructor
}

// NewChain returns a Chain that applies each constructor in order.
func NewChain(constructors ...Constructor) Chain {
	return Chain{append(([]Constructor)(nil), constructors...)}
}

// Then wraps the handler with all middleware in the chain.
func (c Chain) Then(h http.Handler) http.Handler {
	if h == nil {
		h = http.DefaultServeMux
	}
	for i := range c.constructors {
		h = c.constructors[len(c.constructors)-1-i](h)
	}
	return h
}

// ThenFunc is like Then but accepts a handler function.
func (c Chain) ThenFunc(fn http.HandlerFunc) http.Handler {
	if fn == nil {
		return c.Then(nil)
	}
	return c.Then(fn)
}
