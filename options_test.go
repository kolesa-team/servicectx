package xoptions

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestOptionsByService_GetSet(t *testing.T) {
	opts := New()

	require.False(t, opts.HasService("test-service"))
	require.False(t, opts.HasOption("test-service", "test-option"))
	require.Equal(
		t,
		"default-value",
		opts.Get("test-service", "test-option", "default-value"),
		"a default value is expected",
	)

	opts.Set("billing", "branch", "feature-123")
	require.True(t, opts.HasService("billing"))
	require.True(t, opts.HasOption("billing", "branch"))
	require.Equal(
		t,
		"feature-123",
		opts.Get("billing", "branch", "default-branch"),
		"an option must be retrieved after it was set",
	)

	opts.Set("billing", "timeout", "3s")
	require.Equal(
		t,
		time.Second*3,
		opts.GetDuration("billing", "timeout", time.Second),
		"a valid timeout string must be converted to time.Duration",
	)

	opts.Set("billing", "max-value", "100500")
	require.Equal(
		t,
		100500,
		opts.GetInt("billing", "max-value", 0),
		"a valid numeric string must be converted to integer",
	)

	opts.Set("api", "host", "test-host")
	require.Equal(
		t,
		map[string]string{
			"x-service-billing-branch":    "feature-123",
			"x-service-billing-timeout":   "3s",
			"x-service-api-host":          "test-host",
			"x-service-billing-max-value": "100500",
		},
		opts.HeaderMap(),
	)
}

func TestParseServiceHeader(t *testing.T) {
	tests := []struct {
		name            string
		header          string
		wantServiceName string
		wantOption      string
		wantOk          bool
	}{
		{
			name:   "empty header",
			header: "",
			wantOk: false,
		},
		{
			name:   "header with no prefix",
			header: "random-header",
			wantOk: false,
		},
		{
			name:   "incomplete header",
			header: "x-service-abc",
			wantOk: false,
		},
		{
			name:            "valid header",
			header:          "x-service-api-branch",
			wantServiceName: "api",
			wantOption:      "branch",
			wantOk:          true,
		},
		{
			name:            "valid header with a complex option name",
			header:          "x-service-api-timeout-milliseconds",
			wantServiceName: "api",
			wantOption:      "timeout-milliseconds",
			wantOk:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotServiceName, gotOption, gotOk := ParseHeaderString(tt.header)
			if !tt.wantOk {
				require.False(t, gotOk)
				return
			}

			require.Equal(t, tt.wantServiceName, gotServiceName)
			require.Equal(t, tt.wantOption, gotOption)
		})
	}
}
