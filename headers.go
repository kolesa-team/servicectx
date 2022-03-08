package xoptions

import (
	"net/http"
	"strings"
)

// HeaderPrefix Префикс у заголовков межсервисных опций
const HeaderPrefix = "x-service"

// ParseHeaders парсит межсервисные опции из заголовков запроса
func ParseHeaders(headers http.Header) Options {
	result := Options{}

	for name, values := range headers {
		serviceName, option, ok := ParseHeaderString(name)
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

// ParseHeaderString парсит заголовок вида x-service-api-branch
// в название сервиса (api) и название опции (branch)
func ParseHeaderString(header string) (serviceName, option string, ok bool) {
	header = strings.ToLower(header)

	if !strings.HasPrefix(header, HeaderPrefix) {
		return "", "", false
	}

	header = strings.TrimPrefix(header, HeaderPrefix+"-")
	parts := strings.SplitN(header, "-", 2)

	if len(parts) < 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// GetHeaderString возвращает http заголовок для заданной опции и сервиса
func GetHeaderString(serviceName, option string) string {
	return HeaderPrefix + "-" + serviceName + "-" + option
}
