package servicectx

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPropertiesByService_GetSet(t *testing.T) {
	props := New()

	require.False(t, props.HasService("test-service"))
	require.False(t, props.HasProperty("test-service", "test-property"))
	require.Equal(
		t,
		"default-value",
		props.Get("test-service", "test-property", "default-value"),
		"a default value is expected",
	)

	props.Set("billing", "branch", "feature-123")
	require.True(t, props.HasService("billing"))
	require.True(t, props.HasProperty("billing", "branch"))
	require.Equal(
		t,
		"feature-123",
		props.Get("billing", "branch", "default-branch"),
		"an property must be retrieved after it was set",
	)

	props.Set("billing", "timeout", "3s")
	require.Equal(
		t,
		time.Second*3,
		props.GetDuration("billing", "timeout", time.Second),
		"a valid timeout string must be converted to time.Duration",
	)

	props.Set("billing", "max-value", "100500")
	require.Equal(
		t,
		100500,
		props.GetInt("billing", "max-value", 0),
		"a valid numeric string must be converted to integer",
	)

	props.Set("api", "host", "test-host")
	require.Equal(
		t,
		map[string]string{
			"x-service-billing-branch":    "feature-123",
			"x-service-billing-timeout":   "3s",
			"x-service-api-host":          "test-host",
			"x-service-billing-max-value": "100500",
		},
		props.HeaderMap(),
	)
}

func TestParseOptionName(t *testing.T) {
	tests := []struct {
		name            string
		property        string
		wantServiceName string
		wantProperty    string
		wantOk          bool
	}{
		{
			name:     "empty property",
			property: "",
			wantOk:   false,
		},
		{
			name:     "property with no prefix",
			property: "random-property",
			wantOk:   false,
		},
		{
			name:     "incomplete property",
			property: "x-service-abc",
			wantOk:   false,
		},
		{
			name:            "valid property",
			property:        "x-service-api-branch",
			wantServiceName: "api",
			wantProperty:    "branch",
			wantOk:          true,
		},
		{
			name:            "valid property with a complex property name",
			property:        "x-service-api-timeout-milliseconds",
			wantServiceName: "api",
			wantProperty:    "timeout-milliseconds",
			wantOk:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotServiceName, gotOption, gotOk := ParsePropertyName(tt.property)
			if !tt.wantOk {
				require.False(t, gotOk)
				return
			}

			require.Equal(t, tt.wantServiceName, gotServiceName)
			require.Equal(t, tt.wantProperty, gotOption)

			require.Equal(
				t,
				tt.property,
				GetPropertyName(gotServiceName, gotOption),
				"an output of GetPropertyName must be identical to the input of ParsePropertyName",
			)
		})
	}
}

func TestProperties_Merge(t *testing.T) {
	a := New()
	a.Set("a", "version", "1.0")
	a.Set("b", "version", "2.0")

	b := New()
	b.Set("a", "version", "1.1")
	b.Set("b", "branch", "feature-123")

	require.Equal(
		t,
		map[string]string{
			"x-service-a-version": "1.1",
			"x-service-b-version": "2.0",
			"x-service-b-branch":  "feature-123",
		},
		a.Merge(b).HeaderMap(),
	)
}
