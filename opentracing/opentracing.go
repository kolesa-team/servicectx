package opentracing

import (
	"github.com/kolesa-team/xoptions"
	"github.com/opentracing/opentracing-go"
)

func InjectIntoSpan(span opentracing.Span, opts xoptions.Options) {
	for key, value := range opts.HeaderMap() {
		span.SetBaggageItem(key, value)
	}
}

func FromSpan(span opentracing.Span) xoptions.Options {
	opts := xoptions.New()
	span.Context().ForeachBaggageItem(func(key, value string) bool {
		serviceName, option, ok := xoptions.ParseOptionName(key)
		if ok {
			opts.Set(serviceName, option, value)
		}

		return true
	})

	return opts
}
