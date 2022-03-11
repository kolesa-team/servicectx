package opentracing

import (
	"context"
	"github.com/kolesa-team/servicectx"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInjectIntoSpan(t *testing.T) {
	props := servicectx.New()
	props.Set("a", "version", "1.0")
	props.Set("b", "branch", "feature-123")

	span := &mocktracer.MockSpan{}
	InjectIntoSpan(span, props)

	require.Equal(t, "1.0", span.BaggageItem("x-service-a-version"))
	require.Equal(t, "feature-123", span.BaggageItem("x-service-b-branch"))
}

func TestFromSpan(t *testing.T) {
	props := servicectx.New()
	props.Set("a", "version", "1.0")
	props.Set("b", "branch", "feature-123")
	span := &mocktracer.MockSpan{}
	InjectIntoSpan(span, props)

	parsedProps := FromSpan(span)

	require.True(t, parsedProps.HasProperty("a", "version"))
	require.Equal(t, "1.0", parsedProps.Get("a", "version", "9.9"))
	require.True(t, parsedProps.HasProperty("b", "branch"))
	require.Equal(t, "feature-123", parsedProps.Get("b", "branch", "main"))
}

func TestFromContextAndSpan(t *testing.T) {
	// options in span context
	spanProps := servicectx.New()
	spanProps.Set("a", "version", "1.0")
	spanProps.Set("b", "branch", "feature-123")
	span := &mocktracer.MockSpan{}
	InjectIntoSpan(span, spanProps)

	// options in regular Go context (these should have a priority over span context)
	ctxProps := servicectx.New()
	ctxProps.Set("a", "version", "1.1")
	ctx := ctxProps.InjectIntoContext(context.Background())

	parsedProps := FromContextAndSpan(ctx, span)

	require.True(t, parsedProps.HasProperty("a", "version"))
	require.Equal(t, "1.1", parsedProps.Get("a", "version", "9.9"))
	require.True(t, parsedProps.HasProperty("b", "branch"))
	require.Equal(t, "feature-123", parsedProps.Get("b", "branch", "main"))
}
