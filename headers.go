package servicectx

import "net/http"

// FromHeaders constructs options from HTTP option
func FromHeaders(headers http.Header) Options {
	result := Options{}

	for name, values := range headers {
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
