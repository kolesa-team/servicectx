package xoptions

import (
	"github.com/stretchr/testify/require"
	"testing"
)

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
