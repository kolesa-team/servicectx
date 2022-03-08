package example

import (
	"context"
	"fmt"
	"github.com/kolesa-team/xoptions"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

// An example HTTP client and server that use `xoptions` to exchange some parameters through HTTP headers
func ExampleParseHeaders() {
	// a client defines some inter-service options to be sent to a server
	options := xoptions.New()
	options.Set("api", "version", "1.0.1")
	options.Set("billing", "branch", "bugfix-123")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// insert the options into request headers
	options.ApplyToHeaders(req.Header)

	// a simple request handler that parses incoming inter-service options and prints them back to client
	handler := func(w http.ResponseWriter, r *http.Request) {
		options := xoptions.ParseHeaders(r.Header)

		// print an API version, or use a default version 1.0.0
		w.Write([]byte(fmt.Sprintf("API version: %s\n", options.Get("api", "version", "1.0.0"))))

		// configure a billing service URL with a branch from headers, or use a default `main` branch
		billingServiceUrl := xoptions.ReplaceUrlBranch(
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

// An example HTTP client and server where inter-service options are passed through `context.Context`
func ExampleParseHeadersIntoContext() {
	options := xoptions.New()
	options.Set("api", "url", "http://my-custom-api")

	req := httptest.NewRequest(http.MethodGet, "/?username=Alex", nil)
	options.ApplyToHeaders(req.Header)

	handler := func(w http.ResponseWriter, r *http.Request) {
		// parse inter-service options from headers and add them to a context.
		// it's ok if no special headers were sent: an empty struct is then used instead.
		ctx := xoptions.ParseHeadersIntoContext(r.Context(), r.Header)

		// a remoteCall is probably defined in another package;
		// its `username` argument is a part of business logic,
		// but inter-service options are passed in `ctx` as an ancillary data.
		remoteCall := func(ctx context.Context, username string) string {
			// inter-service options are retrieved from a context.
			// the remote API address is taken from these options (or default URL is used instead).
			url := xoptions.FromContext(ctx).Get("api", "url", "http://api")

			url += "?username=" + username
			// TODO: execute remote call
			// http.Get(url)

			return fmt.Sprintf("Remote API url: %s", url)
		}

		w.Write([]byte(remoteCall(ctx, r.URL.Query().Get("username"))))
	}

	w := httptest.NewRecorder()
	handler(w, req)
	res := w.Result()
	responseBytes, _ := ioutil.ReadAll(res.Body)

	fmt.Println(string(responseBytes))

	// Output:
	// Remote API url: http://my-custom-api?username=Alex
}
