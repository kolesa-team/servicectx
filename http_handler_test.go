package servicectx

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

// An example HTTP client and server that exchange properties via HTTP headers
func ExampleFromRequest() {
	// a client defines some properties to be sent to a server
	props := New()
	props.Set("api", "version", "1.0.1")
	props.Set("billing", "branch", "bugfix-123")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// insert the props into request headers
	props.InjectIntoHeaders(req.Header)

	// a simple request handler that parses incoming properties and prints them back in response
	handler := func(w http.ResponseWriter, r *http.Request) {
		props := FromHeaders(r.Header)

		// print an API version, or use a default version 1.0.0
		w.Write([]byte(fmt.Sprintf("API version: %s\n", props.Get("api", "version", "1.0.0"))))

		// configure a billing service URL with a custom branch, or use a default `main` branch
		billingServiceUrl := ReplaceUrlBranch(
			"http://billing-$branch",
			props.Get("billing", "branch", "main"),
		)

		w.Write([]byte(fmt.Sprintf("Billing service url: %s\n", billingServiceUrl)))
	}

	// call the handler above
	w := httptest.NewRecorder()
	handler(w, req)
	res := w.Result()
	responseBytes, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(responseBytes))

	// Output:
	// API version: 1.0.1
	// Billing service url: http://billing-bugfix-123
}

// An example HTTP client and server, with properties exchanged internally through `context.Context`
func ExampleInjectIntoContextFromRequest() {
	props := New()
	// the "api url" will be used by the handler below
	props.Set("api", "url", "http://my-custom-api")
	// and the "billing branch" will just be passed downstream to the remote service
	props.Set("billing", "branch", "hotfix-123")

	req := httptest.NewRequest(http.MethodGet, "/?username=Alex", nil)
	props.InjectIntoHeaders(req.Header)

	handler := func(w http.ResponseWriter, r *http.Request) {
		// parse properties from request and add them to a context.
		// it's ok if no special headers or query args were sent: an empty usable struct is then used instead.
		ctx := InjectIntoContextFromRequest(r.Context(), r)

		// an apiCall is probably defined in another package;
		// its `username` argument is a part of business logic,
		// but properties, being an arbitrary ancillary data, are passed within the context.
		apiCall := func(ctx context.Context, username string) string {
			// props are retrieved from a context
			props := FromContext(ctx)
			// the remote API address is taken from these props (or default URL is used instead).
			url := props.Get("api", "url", "http://api")
			url += "?username=" + username
			apiRequest, _ := http.NewRequest("GET", url, nil)
			// the properties are propagated further within the headers
			props.InjectIntoHeaders(apiRequest.Header)

			// ...execute remote call
			// _, _ = http.DefaultClient.Do(apiRequest)

			return fmt.Sprintf("Calling remote API at %s with headers:\n%+v", url, apiRequest.Header)
		}

		// execute remote call and print the results back
		w.Write([]byte(apiCall(ctx, r.URL.Query().Get("username"))))
	}

	w := httptest.NewRecorder()
	handler(w, req)
	res := w.Result()
	responseBytes, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(responseBytes))

	// Output:
	// Calling remote API at http://my-custom-api?username=Alex with headers:
	// map[X-Service-Api-Url:[http://my-custom-api] X-Service-Billing-Branch:[hotfix-123]]
}
