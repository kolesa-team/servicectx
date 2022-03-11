package opentracing

import (
	"context"
	"github.com/kolesa-team/servicectx"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInjectIntoSpan(t *testing.T) {
	opts := servicectx.New()
	opts.Set("a", "version", "1.0")
	opts.Set("b", "branch", "feature-123")

	span := &mocktracer.MockSpan{}
	InjectIntoSpan(span, opts)

	require.Equal(t, "1.0", span.BaggageItem("x-service-a-version"))
	require.Equal(t, "feature-123", span.BaggageItem("x-service-b-branch"))
}

func TestFromSpan(t *testing.T) {
	inputOpts := servicectx.New()
	inputOpts.Set("a", "version", "1.0")
	inputOpts.Set("b", "branch", "feature-123")
	span := &mocktracer.MockSpan{}
	InjectIntoSpan(span, inputOpts)

	parsedOpts := FromSpan(span)

	require.True(t, parsedOpts.HasOption("a", "version"))
	require.Equal(t, "1.0", parsedOpts.Get("a", "version", "9.9"))
	require.True(t, parsedOpts.HasOption("b", "branch"))
	require.Equal(t, "feature-123", parsedOpts.Get("b", "branch", "main"))
}

func TestFromContextAndSpan(t *testing.T) {
	// options in span context
	spanOpts := servicectx.New()
	spanOpts.Set("a", "version", "1.0")
	spanOpts.Set("b", "branch", "feature-123")
	span := &mocktracer.MockSpan{}
	InjectIntoSpan(span, spanOpts)

	// options in regular Go context (these should have a priority over span context)
	ctxOpts := servicectx.New()
	ctxOpts.Set("a", "version", "1.1")
	ctx := ctxOpts.InjectIntoContext(context.Background())

	parsedOpts := FromContextAndSpan(ctx, span)

	require.True(t, parsedOpts.HasOption("a", "version"))
	require.Equal(t, "1.1", parsedOpts.Get("a", "version", "9.9"))
	require.True(t, parsedOpts.HasOption("b", "branch"))
	require.Equal(t, "feature-123", parsedOpts.Get("b", "branch", "main"))
}
