package servicectx

import (
	"context"
	"net/http"
)

// Utility functions for passing properties between `http.Header` and `context.Context`

type contextKey string

const contextKeyOptions = contextKey("servicectx")

// FromContext returns properties from context.
// If there are no properties in the context, an empty usable instance is returned.
func FromContext(ctx context.Context) Properties {
	props, ok := ctx.Value(contextKeyOptions).(Properties)
	if ok {
		return props
	}

	return New()
}

// InjectIntoContext adds properties to the context
func (p Properties) InjectIntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyOptions, p)
}

// InjectIntoContextFromRequest parses properties from request and adds them into context
func InjectIntoContextFromRequest(ctx context.Context, req *http.Request) context.Context {
	return FromRequest(req).InjectIntoContext(ctx)
}

// InjectIntoHeadersFromContext adds properties from context into http.Header
func InjectIntoHeadersFromContext(ctx context.Context, header http.Header) {
	FromContext(ctx).InjectIntoHeaders(header)
}
