// Package xoptions allows to easily exchange arbitrary options across microservices via HTTP headers.
// It handles parsing, reading, and writing the options to/from http.Header, and passing them through `context`.
package xoptions

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Values a key => value options
type Values map[string]string

// Options grouped by a service name
type Options map[string]Values

// NamePrefix a prefix at the beginning of an inter-service option
const NamePrefix = "x-service"

// ParseOptionName parses a string like "x-service-api-branch" into service name ("api"),
// option name ("branch"), and a boolean success flag
func ParseOptionName(name string) (serviceName, option string, ok bool) {
	name = strings.ToLower(name)

	if !strings.HasPrefix(name, NamePrefix) {
		return "", "", false
	}

	name = strings.TrimPrefix(name, NamePrefix+"-")
	parts := strings.SplitN(name, "-", 2)

	if len(parts) < 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// GetOptionName builds a string from service name and option name
func GetOptionName(serviceName, option string) string {
	return NamePrefix + "-" + serviceName + "-" + option
}

// New constructs a new options instance
func New() Options {
	return Options{}
}

// HasService checks if there are options for a given service
func (opts Options) HasService(serviceName string) bool {
	_, ok := opts[serviceName]
	return ok
}

// HasOption checks if a given option exists for a service
func (opts Options) HasOption(serviceName, option string) bool {
	if service, ok := opts[serviceName]; ok {
		if _, ok := service[option]; ok {
			return true
		}
	}

	return false
}

// GetByService returns all options for a given service
func (opts Options) GetByService(serviceName string) Values {
	return opts[serviceName]
}

// Get returns an option value for a given service
func (opts Options) Get(serviceName, option, defaultValue string) string {
	if service, ok := opts[serviceName]; ok {
		if value, ok := service[option]; ok {
			return value
		}
	}

	return defaultValue
}

// GetInt returns an option value for a given service as an integer
func (opts Options) GetInt(serviceName, option string, defaultValue int) int {
	valueStr := opts.Get(serviceName, option, "")
	if len(valueStr) == 0 {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetDuration returns an option value for a given service as time.Duration
func (opts Options) GetDuration(serviceName, option string, defaultValue time.Duration) time.Duration {
	valueStr := opts.Get(serviceName, option, "")
	if len(valueStr) == 0 {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// Set sets an option value for a given service
func (opts Options) Set(serviceName, option, value string) Options {
	if _, ok := opts[serviceName]; !ok {
		opts[serviceName] = map[string]string{}
	}

	opts[serviceName][option] = value

	return opts
}

// HeaderMap returns options as a map of HTTP headers
func (opts Options) HeaderMap() map[string]string {
	result := map[string]string{}

	for service, options := range opts {
		for key, value := range options {
			result[GetOptionName(service, key)] = value
		}
	}

	return result
}

// InjectIntoHeaders adds option headers to http.Header
func (opts Options) InjectIntoHeaders(headers http.Header) {
	for name, value := range opts.HeaderMap() {
		headers.Set(name, value)
	}
}
