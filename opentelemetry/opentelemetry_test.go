package opentelemetry

import (
	"context"
	"fmt"
	"github.com/kolesa-team/xoptions"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"net/http"
	"testing"
)

func TestInjectIntoBaggage(t *testing.T) {
	bag := baggage.Baggage{}
	opts := xoptions.New()
	opts.Set("a", "version", "1.0")
	opts.Set("b", "branch", "feature-123")

	bag = InjectIntoBaggage(bag, opts)

	require.Equal(t, "1.0", bag.Member("x-service-a-version").Value())
	require.Equal(t, "feature-123", bag.Member("x-service-b-branch").Value())
}

func TestFromBaggage(t *testing.T) {
	bag := baggage.Baggage{}
	inputOpts := xoptions.New()
	inputOpts.Set("a", "version", "1.0")
	inputOpts.Set("b", "branch", "feature-123")
	bag = InjectIntoBaggage(bag, inputOpts)

	parsedOpts := FromBaggage(bag)
	require.True(t, parsedOpts.HasOption("a", "version"))
	require.Equal(t, "1.0", parsedOpts.Get("a", "version", "9.9"))
	require.True(t, parsedOpts.HasOption("b", "branch"))
	require.Equal(t, "feature-123", parsedOpts.Get("b", "branch", "main"))
}

func ExampleFromBaggage() {
	prop := propagation.TextMapPropagator(propagation.Baggage{})
	req, _ := http.NewRequest("GET", "http://opentelemetry.com", nil)
	req.Header.Set("baggage", "x-service-a-version=2.0,x-service-b-branch=bugfix-123")

	ctx := context.Background()
	ctx = prop.Extract(ctx, propagation.HeaderCarrier(req.Header))
	bag := baggage.FromContext(ctx)
	opts := FromBaggage(bag)

	fmt.Println("service A version:", opts.Get("a", "version", ""))
	fmt.Println("service B branch:", opts.Get("b", "branch", ""))

	// Output:
	// service A version: 2.0
	// service B branch: bugfix-123
}

func ExampleInjectIntoBaggage() {
	inputOpts := xoptions.New()
	inputOpts.Set("a", "version", "1.0")
	inputOpts.Set("b", "branch", "feature-123")

	propagator := propagation.Baggage{}
	bag := InjectIntoBaggage(baggage.Baggage{}, inputOpts)
	ctx := baggage.ContextWithBaggage(context.Background(), bag)
	req, _ := http.NewRequest("GET", "http://opentelemetry.com", nil)
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	fmt.Println(req.Header.Get("baggage"))
	// Output:
	// x-service-a-version=1.0,x-service-b-branch=feature-123
}
