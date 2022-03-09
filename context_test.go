package xoptions

import (
	"context"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestFromContext(t *testing.T) {
	emptyCtx := context.Background()
	require.Equal(
		t,
		Options{},
		FromContext(emptyCtx),
		"an empty options struct must be returned from an empty context",
	)
}

func TestOptions_InjectIntoContext(t *testing.T) {
	opts := New()
	opts.Set("a", "option", "value-a")
	opts.Set("b", "option", "value-b")

	ctx := opts.InjectIntoContext(context.Background())

	require.Equal(
		t,
		Options{
			"a": Values{
				"option": "value-a",
			},
			"b": Values{
				"option": "value-b",
			},
		},
		FromContext(ctx),
		"options must be successfully added to and retrieved from the context",
	)
}

func Test_InjectIntoHeadersFromContext(t *testing.T) {
	opts := New()
	opts.Set("a", "option", "value-a")
	opts.Set("b", "option", "value-b")
	ctx := opts.InjectIntoContext(context.Background())

	httpHeader := http.Header{}
	InjectIntoHeadersFromContext(ctx, httpHeader)

	require.Equal(t, "value-a", httpHeader.Get("x-service-a-option"))
	require.Equal(t, "value-b", httpHeader.Get("x-service-b-option"))
}

func Test_InjectIntoContextFromHeaders(t *testing.T) {
	httpHeader := http.Header{}
	httpHeader.Set("X-Service-A-Option", "value-a")
	httpHeader.Set("x-service-b-option", "value-b")

	ctx := InjectIntoContextFromHeaders(context.Background(), httpHeader)
	require.Equal(
		t,
		Options{
			"a": Values{
				"option": "value-a",
			},
			"b": Values{
				"option": "value-b",
			},
		},
		FromContext(ctx),
		"options must be parsed from http.Header",
	)
}
