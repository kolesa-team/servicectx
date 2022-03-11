package otel

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

// An example HTTP client and server exchanging custom properties through opentelemetry baggage
func Example() {
	// a client sends a request with a custom API url;
	// this property is passed in HTTP headers as part of an opentelemetry baggage
	props := servicectx.New()
	props.Set("api", "url", "http://my-custom-api")

	propagator := propagation.TextMapPropagator(propagation.Baggage{})
	ctx := InjectIntoContext(context.Background(), props)
	req := httptest.NewRequest(http.MethodGet, "/?username=Alex", nil)
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// a server handles this request, modifies the API url accordingly,
	// and then also passes these custom properties in a remote call
	handler := func(w http.ResponseWriter, r *http.Request) {
		// parse the properties from opentelemetry baggage and add them to a Go context for use in application code
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		bag := baggage.FromContext(ctx)
		ctx = FromBaggage(bag).InjectIntoContext(ctx)

		// an apiCall is probably defined in another package;
		// its `username` argument is a part of business logic,
		// but properties, being an arbitrary ancillary data, are passed within the context.
		apiCall := func(ctx context.Context, username string) string {
			props := servicectx.FromContext(ctx)
			// the remote API address is taken from these properties (or default URL is used instead).
			url := props.Get("api", "url", "http://api")
			url += "?username=" + username
			apiRequest, _ := http.NewRequest("GET", url, nil)

			// the props are propagated further through opentelemetry baggage
			propagator.Inject(ctx, propagation.HeaderCarrier(apiRequest.Header))

			// ...execute remote call
			// _, _ = http.DefaultClient.Do(apiRequest)

			return fmt.Sprintf("Calling remote API at %s with baggage: %s", url, apiRequest.Header.Get("baggage"))
		}

		apiCallResult := apiCall(ctx, r.URL.Query().Get("username"))
		w.Write([]byte(apiCallResult))
	}

	w := httptest.NewRecorder()
	handler(w, req)
	res := w.Result()
	responseBytes, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(responseBytes))

	// Output:
	// Calling remote API at http://my-custom-api?username=Alex with baggage: x-service-api-url=http%3A%2F%2Fmy-custom-api
}
