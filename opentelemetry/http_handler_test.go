package opentelemetry

import (
	"context"
	"fmt"
	"github.com/kolesa-team/servicectx"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

// An an example HTTP client and server exchanging inter-service options through opentelemetry baggage
func ExampleBaggage() {
	options := servicectx.New()
	options.Set("api", "url", "http://my-custom-api")

	propagator := propagation.TextMapPropagator(propagation.Baggage{})
	bag := InjectIntoBaggage(baggage.Baggage{}, options)
	ctx := baggage.ContextWithBaggage(context.Background(), bag)

	req := httptest.NewRequest(http.MethodGet, "/?username=Alex", nil)
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	handler := func(w http.ResponseWriter, r *http.Request) {
		// parse inter-service options from opentracing baggage and add them to a context.
		ctx = propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		bag := baggage.FromContext(ctx)
		ctx = FromBaggage(bag).InjectIntoContext(ctx)

		// a remoteCall is probably defined in another package;
		// its `username` argument is a part of business logic,
		// but inter-service options are passed in `ctx` as an ancillary data.
		remoteCall := func(ctx context.Context, username string) string {
			// inter-service options are retrieved from a context
			opts := servicectx.FromContext(ctx)
			// the remote API address is taken from these options (or default URL is used instead).
			url := opts.Get("api", "url", "http://api")
			url += "?username=" + username
			apiRequest, _ := http.NewRequest("GET", url, nil)

			// the options are propagated further through opentelemetry baggage
			bag := baggage.FromContext(ctx)
			bag = InjectIntoBaggage(bag, opts)
			ctx = baggage.ContextWithBaggage(ctx, bag)
			propagator.Inject(ctx, propagation.HeaderCarrier(apiRequest.Header))

			// TODO: execute remote call
			// _, _ = http.DefaultClient.Do(apiRequest)

			return fmt.Sprintf("Calling remote API at %s with baggage %s", url, apiRequest.Header.Get("baggage"))
		}

		w.Write([]byte(remoteCall(ctx, r.URL.Query().Get("username"))))
	}

	w := httptest.NewRecorder()
	handler(w, req)
	res := w.Result()
	responseBytes, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(responseBytes))

	// Output:
	// Calling remote API at http://my-custom-api?username=Alex with baggage x-service-api-url=http%3A%2F%2Fmy-custom-api
}
