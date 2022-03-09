package xoptions

import (
	"context"
	"net/http"
)

// Utility functions for passing inter-service options between `http.Header` and `context.Context`

type contextKey string

const contextKeyOptions = contextKey("xoptions")

// FromContext returns options from context.
// If there are no options in the context, an empty struct is returned.
func FromContext(ctx context.Context) Options {
	options, ok := ctx.Value(contextKeyOptions).(Options)
	if ok {
		return options
	}

	return New()
}

// InjectIntoContext adds options to the context
func (opts Options) InjectIntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyOptions, opts)
}

// InjectIntoContextFromHeaders parses options from http.Header and adds them into context
func InjectIntoContextFromHeaders(ctx context.Context, headers http.Header) context.Context {
	return FromHeaders(headers).InjectIntoContext(ctx)
}

// InjectIntoHeadersFromContext adds options from context into http.Header
func InjectIntoHeadersFromContext(ctx context.Context, header http.Header) {
	FromContext(ctx).InjectIntoHeaders(header)
}
