package xoptions

import (
	"net/url"
)

// QueryString converts options to an HTTP query string
func (opts Options) QueryString() string {
	values := url.Values{}

	for service, options := range opts {
		for key, value := range options {
			values.Set(GetOptionName(service, key), value)
		}
	}

	return values.Encode()
}

// FromQueryString parses options from an HTTP query string
func FromQueryString(query string) Options {
	result := Options{}
	parsedQuery, err := url.ParseQuery(query)
	if err != nil {
		return result
	}

	for name, values := range parsedQuery {
		serviceName, option, ok := ParseOptionName(name)
		if !ok {
			continue
		}

		if _, ok := result[serviceName]; !ok {
			result[serviceName] = map[string]string{}
		}

		result[serviceName][option] = values[0]
	}

	return result
}
