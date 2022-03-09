package xoptions

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOptions_QueryString(t *testing.T) {
	opts := New()
	require.Equal(t, "", opts.QueryString())

	opts.Set("a", "version", "1.2")
	require.Equal(t, "x-service-a-version=1.2", opts.QueryString())

	opts.Set("b", "branch", "my-*special*-branch")
	require.Equal(t, "x-service-a-version=1.2&x-service-b-branch=my-%2Aspecial%2A-branch", opts.QueryString())
}

func TestFromQueryString(t *testing.T) {
	require.Empty(t, FromQueryString(""))
	require.Empty(t, FromQueryString("city=Almaty&country=Kazakhstan"))

	require.Equal(
		t,
		New().Set("a", "version", "1.2"),
		FromQueryString("x-service-a-version=1.2"),
	)

	require.Equal(
		t,
		New().Set("a", "version", "1.2").
			Set("b", "branch", "my-*special*-branch"),
		FromQueryString("x-service-a-version=1.2&x-service-b-branch=my-%2Aspecial%2A-branch"),
	)
}
