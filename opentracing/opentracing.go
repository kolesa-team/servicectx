package opentracing

import (
	"github.com/kolesa-team/servicectx"
	"github.com/opentracing/opentracing-go"
)

func InjectIntoSpan(span opentracing.Span, opts servicectx.Options) {
	for key, value := range opts.HeaderMap() {
		span.SetBaggageItem(key, value)
	}
}

func FromSpan(span opentracing.Span) servicectx.Options {
	return FromSpanContext(span.Context())
}

func FromSpanContext(spanCtx opentracing.SpanContext) servicectx.Options {
	opts := servicectx.New()
	spanCtx.ForeachBaggageItem(func(key, value string) bool {
		serviceName, option, ok := servicectx.ParseOptionName(key)
		if ok {
			opts.Set(serviceName, option, value)
		}

		return true
	})

	return opts
}
