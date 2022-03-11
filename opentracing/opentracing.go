package opentracing

import (
	"context"
	"github.com/kolesa-team/servicectx"
	"github.com/opentracing/opentracing-go"
)

// InjectIntoSpan adds properties into span's baggage
func InjectIntoSpan(span opentracing.Span, props servicectx.Properties) {
	for key, value := range props.HeaderMap() {
		span.SetBaggageItem(key, value)
	}
}

// FromContextAndSpan retrieves properties from Go context and from span's context.
// This is convenient when the properties can be set both in application code via context and from outside world by opentracing.
// The properties from Go context have a preference over span's context.
func FromContextAndSpan(ctx context.Context, span opentracing.Span) servicectx.Properties {
	return FromSpan(span).Merge(servicectx.FromContext(ctx))
}

// FromSpan retrieves properties from span
func FromSpan(span opentracing.Span) servicectx.Properties {
	return FromSpanContext(span.Context())
}

// FromSpanContext retrieves properties from span's context
func FromSpanContext(spanCtx opentracing.SpanContext) servicectx.Properties {
	props := servicectx.New()
	spanCtx.ForeachBaggageItem(func(key, value string) bool {
		serviceName, option, ok := servicectx.ParsePropertyName(key)
		if ok {
			props.Set(serviceName, option, value)
		}

		return true
	})

	return props
}
