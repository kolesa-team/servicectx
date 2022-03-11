package otel

import (
	"context"
	"fmt"
	"github.com/kolesa-team/servicectx"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"net/http"
	"testing"
)

func TestInjectIntoBaggage(t *testing.T) {
	bag := baggage.Baggage{}
	props := servicectx.New()
	props.Set("a", "version", "1.0")
	props.Set("b", "branch", "feature-123")

	bag = InjectIntoBaggage(bag, props)

	require.Equal(t, "1.0", bag.Member("x-service-a-version").Value())
	require.Equal(t, "feature-123", bag.Member("x-service-b-branch").Value())
}

func TestFromBaggage(t *testing.T) {
	bag := baggage.Baggage{}
	props := servicectx.New()
	props.Set("a", "version", "1.0")
	props.Set("b", "branch", "feature-123")
	bag = InjectIntoBaggage(bag, props)

	parsedProps := FromBaggage(bag)
	require.True(t, parsedProps.HasProperty("a", "version"))
	require.Equal(t, "1.0", parsedProps.Get("a", "version", "9.9"))
	require.True(t, parsedProps.HasProperty("b", "branch"))
	require.Equal(t, "feature-123", parsedProps.Get("b", "branch", "main"))
}

func ExampleFromBaggage() {
	prop := propagation.TextMapPropagator(propagation.Baggage{})
	req, _ := http.NewRequest("GET", "http://opentelemetry.com", nil)
	req.Header.Set("baggage", "x-service-a-version=2.0,x-service-b-branch=bugfix-123")

	ctx := context.Background()
	ctx = prop.Extract(ctx, propagation.HeaderCarrier(req.Header))
	bag := baggage.FromContext(ctx)
	props := FromBaggage(bag)

	fmt.Println("service A version:", props.Get("a", "version", ""))
	fmt.Println("service B branch:", props.Get("b", "branch", ""))

	// Output:
	// service A version: 2.0
	// service B branch: bugfix-123
}

func ExampleInjectIntoBaggage() {
	props := servicectx.New()
	props.Set("a", "version", "1.0")
	props.Set("b", "branch", "feature-123")

	propagator := propagation.Baggage{}
	bag := InjectIntoBaggage(baggage.Baggage{}, props)
	ctx := baggage.ContextWithBaggage(context.Background(), bag)
	req, _ := http.NewRequest("GET", "http://opentelemetry.com", nil)
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	fmt.Println(req.Header.Get("baggage"))
	// Output:
	// x-service-a-version=1.0,x-service-b-branch=feature-123
}

func TestFromContextAndBaggage(t *testing.T) {
	// options in baggage
	bag := baggage.Baggage{}
	inputProps := servicectx.New()
	inputProps.Set("a", "version", "1.0")
	inputProps.Set("b", "branch", "feature-123")
	bag = InjectIntoBaggage(bag, inputProps)

	// options in regular Go context (these should have a priority over baggage)
	ctxProps := servicectx.New()
	ctxProps.Set("a", "version", "1.1")
	ctx := ctxProps.InjectIntoContext(context.Background())

	parsedProps := FromContextAndBaggage(ctx, bag)

	require.True(t, parsedProps.HasProperty("a", "version"))
	require.Equal(t, "1.1", parsedProps.Get("a", "version", "9.9"))
	require.True(t, parsedProps.HasProperty("b", "branch"))
	require.Equal(t, "feature-123", parsedProps.Get("b", "branch", "main"))
}

func TestInjectIntoContext(t *testing.T) {
	props := servicectx.New()
	props.Set("a", "version", "1.0")
	props.Set("b", "branch", "feature-123")

	ctx := InjectIntoContext(context.Background(), props)

	bag := baggage.FromContext(ctx)
	require.Equal(t, "1.0", bag.Member("x-service-a-version").Value())
	require.Equal(t, "feature-123", bag.Member("x-service-b-branch").Value())

	// ensure that multiple calls to InjectIntoContext result in properties merged correctly
	props2 := servicectx.New()
	props2.Set("a", "version", "2.0")
	props2.Set("c", "timeout", "3s")

	ctx = InjectIntoContext(ctx, props2)
	bag = baggage.FromContext(ctx)
	require.Equal(t, "2.0", bag.Member("x-service-a-version").Value())
	require.Equal(t, "feature-123", bag.Member("x-service-b-branch").Value())
	require.Equal(t, "3s", bag.Member("x-service-c-timeout").Value())
}
