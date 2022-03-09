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
	opts := servicectx.New()
	span.Context().ForeachBaggageItem(func(key, value string) bool {
		serviceName, option, ok := servicectx.ParseOptionName(key)
		if ok {
			opts.Set(serviceName, option, value)
		}

		return true
	})

	return opts
}
