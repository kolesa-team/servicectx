package servicectx

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
		Properties{},
		FromContext(emptyCtx),
		"an empty struct must be returned from an empty context",
	)
}

func TestProperties_InjectIntoContext(t *testing.T) {
	props := New()
	props.Set("a", "property", "value-a")
	props.Set("b", "property", "value-b")

	ctx := props.InjectIntoContext(context.Background())

	require.Equal(
		t,
		Properties{
			"a": Values{
				"property": "value-a",
			},
			"b": Values{
				"property": "value-b",
			},
		},
		FromContext(ctx),
		"properties must be successfully added to and retrieved from the context",
	)
}

func Test_InjectIntoHeadersFromContext(t *testing.T) {
	props := New()
	props.Set("a", "property", "value-a")
	props.Set("b", "property", "value-b")
	ctx := props.InjectIntoContext(context.Background())

	httpHeader := http.Header{}
	InjectIntoHeadersFromContext(ctx, httpHeader)

	require.Equal(t, "value-a", httpHeader.Get("x-service-a-property"))
	require.Equal(t, "value-b", httpHeader.Get("x-service-b-property"))
}

func Test_InjectIntoContextFromRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Service-A-Option", "value-a")
	req.Header.Set("x-service-b-option", "value-b")
	req.URL.RawQuery = "x-service-c-option=value-c"

	ctx := InjectIntoContextFromRequest(context.Background(), req)
	require.Equal(
		t,
		Properties{
			"a": Values{
				"option": "value-a",
			},
			"b": Values{
				"option": "value-b",
			},
			"c": Values{
				"option": "value-c",
			},
		},
		FromContext(ctx),
		"properties must be parsed from http.Header",
	)
}
