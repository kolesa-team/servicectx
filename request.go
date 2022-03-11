package servicectx

import (
	"net/http"
	"net/url"
	"strings"
)

// QueryString converts properties to an HTTP query string
func (p Properties) QueryString() string {
	return p.QueryValues().Encode()
}

// QueryValues converts properties to a set of HTTP query parameters
func (p Properties) QueryValues() url.Values {
	values := url.Values{}

	for service, props := range p {
		for key, value := range props {
			values.Set(GetPropertyName(service, key), value)
		}
	}

	return values
}

// FromQueryString parses properties from an HTTP query string
func FromQueryString(query string) Properties {
	parsedQuery, err := url.ParseQuery(query)
	if err != nil {
		return New()
	}

	return FromQueryValues(parsedQuery)
}

// FromQueryValues parses properties from a parsed HTTP query string
func FromQueryValues(values url.Values) Properties {
	props := Properties{}

	for name, values := range values {
		serviceName, option, ok := ParsePropertyName(name)
		if !ok {
			continue
		}

		if _, ok := props[serviceName]; !ok {
			props[serviceName] = map[string]string{}
		}

		props[serviceName][option] = values[0]
	}

	return props
}

// FromHeaders constructs properties from HTTP headers
func FromHeaders(headers http.Header) Properties {
	props := New()

	for name, values := range headers {
		serviceName, option, ok := ParsePropertyName(name)
		if !ok {
			continue
		}

		if _, ok := props[serviceName]; !ok {
			props[serviceName] = map[string]string{}
		}

		props[serviceName][option] = values[0]
	}

	return props
}

// FromRequest constructs properties from HTTP headers and query string of the request.
// Query string properties have a priority over HTTP headers.
func FromRequest(req *http.Request) Properties {
	fromHeaders := FromHeaders(req.Header)
	fromQuery := FromQueryValues(req.URL.Query())

	return fromHeaders.Merge(fromQuery)
}

// UrlBranchPlaceholder is a part of URL to be replaced with a branch name
const UrlBranchPlaceholder = "$branch"

// ReplaceUrlBranch replaces branch placeholder in URL with an actual branch name.
func ReplaceUrlBranch(url, branch string) string {
	if branch == "" || !strings.Contains(url, UrlBranchPlaceholder) {
		return url
	}

	// try to normalize branch name for safe use in URLs
	branch = strings.ToLower(strings.TrimSpace(branch))
	branch = strings.ReplaceAll(branch, "_", "-")

	return strings.ReplaceAll(url, UrlBranchPlaceholder, branch)
}
