// Package servicectx facilitates exchanging arbitrary properties across microservices via HTTP headers, query strings,
// OpenTelemetry/OpenTracing baggage, and/or Go Context.
package servicectx

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Values key => value properties
type Values map[string]string

// Properties grouped by a service name
type Properties map[string]Values

// NamePrefix a prefix at the beginning of a property name indicating it belongs to this package
const NamePrefix = "x-service"
const Separator = "-"

// ParsePropertyName parses a string like "x-service-api-branch" into service name ("api"),
// property name ("branch"), and a boolean success flag
func ParsePropertyName(name string) (serviceName, option string, ok bool) {
	name = strings.ToLower(name)

	if !strings.HasPrefix(name, NamePrefix) {
		return "", "", false
	}

	name = strings.TrimPrefix(name, NamePrefix+Separator)
	parts := strings.SplitN(name, Separator, 2)

	if len(parts) < 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// GetPropertyName builds a string from service name and property name
func GetPropertyName(serviceName, option string) string {
	return NamePrefix + Separator + serviceName + Separator + option
}

// New constructs a new properties instance
func New() Properties {
	return Properties{}
}

// HasService checks if there are options for a given service
func (p Properties) HasService(serviceName string) bool {
	_, ok := p[serviceName]
	return ok
}

// HasProperty checks if a given property exists for a service
func (p Properties) HasProperty(serviceName, option string) bool {
	if service, ok := p[serviceName]; ok {
		if _, ok := service[option]; ok {
			return true
		}
	}

	return false
}

// GetByService returns all options for a given service
func (p Properties) GetByService(serviceName string) Values {
	return p[serviceName]
}

// Get returns an property value for a given service
func (p Properties) Get(serviceName, prop, defaultValue string) string {
	if service, ok := p[serviceName]; ok {
		if value, ok := service[prop]; ok {
			return value
		}
	}

	return defaultValue
}

// GetInt returns a property value for a given service as an integer
func (p Properties) GetInt(serviceName, prop string, defaultValue int) int {
	valueStr := p.Get(serviceName, prop, "")
	if len(valueStr) == 0 {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetDuration returns a property value for a given service as time.Duration
func (p Properties) GetDuration(serviceName, prop string, defaultValue time.Duration) time.Duration {
	valueStr := p.Get(serviceName, prop, "")
	if len(valueStr) == 0 {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetBool returns a property value for a given service as boolean
func (p Properties) GetBool(serviceName, prop string, defaultValue bool) bool {
	valueStr := p.Get(serviceName, prop, "")
	if len(valueStr) == 0 {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// Set sets a property value for a given service
func (p Properties) Set(serviceName, prop, value string) Properties {
	if _, ok := p[serviceName]; !ok {
		p[serviceName] = map[string]string{}
	}

	p[serviceName][prop] = value

	return p
}

// HeaderMap returns options as a map of HTTP headers
func (p Properties) HeaderMap() map[string]string {
	result := map[string]string{}

	for service, props := range p {
		for key, value := range props {
			result[GetPropertyName(service, key)] = value
		}
	}

	return result
}

// InjectIntoHeaders adds property headers to http.Header
func (p Properties) InjectIntoHeaders(headers http.Header) {
	for name, value := range p.HeaderMap() {
		headers.Set(name, value)
	}
}

// Merge merges two sets of properties. The receiver is modified and returned for chaining.
func (p Properties) Merge(other Properties) Properties {
	for serviceName, values := range other {
		for key, value := range values {
			p.Set(serviceName, key, value)
		}
	}

	return p
}
