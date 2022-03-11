package servicectx

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestProperties_QueryString(t *testing.T) {
	props := New()
	require.Equal(t, "", props.QueryString())

	props.Set("a", "version", "1.2")
	require.Equal(t, "x-service-a-version=1.2", props.QueryString())

	props.Set("b", "branch", "my-*special*-branch")
	require.Equal(t, "x-service-a-version=1.2&x-service-b-branch=my-%2Aspecial%2A-branch", props.QueryString())
}

func TestFromQueryString(t *testing.T) {
	require.Empty(t, FromQueryString(""))
	require.Empty(t, FromQueryString("invalid;query"))
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

func TestReplaceUrlBranch(t *testing.T) {
	require.Equal(
		t,
		"test-host",
		ReplaceUrlBranch("test-host", ""),
		"url should not be changed if it doesn't contain the placeholder and the branch is empty",
	)

	require.Equal(
		t,
		"test-host",
		ReplaceUrlBranch("test-host", "test-branch"),
		"url should not be changed if it doesn't contain the placeholder",
	)

	require.Equal(
		t,
		"test-host-feature-123",
		ReplaceUrlBranch("test-host-$branch", "feature-123"),
		"url should be changed if the placeholder and branch are not empty",
	)

	require.Equal(
		t,
		"test-host-feature-123",
		ReplaceUrlBranch("test-host-$branch", "\tFeature_123   "),
		"branch name should be normalized",
	)
}

func TestFromRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "test-url", nil)
	req.Header.Set("unknown-header", "unknown-value")
	req.Header.Set("x-service-api-branch", "feature-123")
	req.Header.Set("x-service-billing-version", "1.2")

	query := req.URL.Query()
	query.Set("x-service-billing-version", "2.2")
	query.Set("unknown-param", "unknown-value")
	req.URL.RawQuery = query.Encode()

	props := FromRequest(req)
	require.Equal(t, "feature-123", props.Get("api", "branch", "main"))
	require.Equal(t, "2.2", props.Get("billing", "version", "1.0"))
}
