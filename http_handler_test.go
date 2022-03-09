package servicectx

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

// An example HTTP client and server that exchange inter-service options through HTTP headers
func ExampleFromHeaders() {
	// a client defines some options to be sent to a server
	options := New()
	options.Set("api", "version", "1.0.1")
	options.Set("billing", "branch", "bugfix-123")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// insert the options into request headers
	options.InjectIntoHeaders(req.Header)

	// a simple request handler that parses incoming inter-service options and prints them back to client
	handler := func(w http.ResponseWriter, r *http.Request) {
		options := FromHeaders(r.Header)

		// print an API version, or use a default version 1.0.0
		w.Write([]byte(fmt.Sprintf("API version: %s\n", options.Get("api", "version", "1.0.0"))))

		// configure a billing service URL with a branch from headers, or use a default `main` branch
		billingServiceUrl := ReplaceUrlBranch(
			"http://billing-$branch",
			options.Get("billing", "branch", "main"),
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

// An example HTTP client and server where options are passed through `context.Context`
func ExampleInjectIntoContextFromHeaders() {
	options := New()
	// the "api url" will be used by the handler below
	options.Set("api", "url", "http://my-custom-api")
	// and the "billing branch" will just be passed downstream to the remote service
	options.Set("billing", "branch", "hotfix-123")

	req := httptest.NewRequest(http.MethodGet, "/?username=Alex", nil)
	options.InjectIntoHeaders(req.Header)

	handler := func(w http.ResponseWriter, r *http.Request) {
		// parse inter-service options from headers and add them to a context.
		// it's ok if no special headers were sent: an empty struct is then used instead.
		ctx := InjectIntoContextFromHeaders(r.Context(), r.Header)

		// a remoteCall is probably defined in another package;
		// its `username` argument is a part of business logic,
		// but inter-service options are passed in `ctx` as an ancillary data.
		remoteCall := func(ctx context.Context, username string) string {
			// options are retrieved from a context
			opts := FromContext(ctx)
			// the remote API address is taken from these options (or default URL is used instead).
			url := opts.Get("api", "url", "http://api")
			url += "?username=" + username
			apiRequest, _ := http.NewRequest("GET", url, nil)
			// the options are propagated further within the headers
			opts.InjectIntoHeaders(apiRequest.Header)
			// TODO: execute remote call
			// _, _ = http.DefaultClient.Do(apiRequest)

			return fmt.Sprintf("Calling remote API at %s with headers:\n%+v", url, apiRequest.Header)
		}

		w.Write([]byte(remoteCall(ctx, r.URL.Query().Get("username"))))
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
