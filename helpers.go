package servicectx

import "strings"

// BranchPlaceholder is a part of URL to be replaced with a branch name
const BranchPlaceholder = "$branch"

// ReplaceUrlBranch replaces branch placeholder in URL with an actual branch name.
func ReplaceUrlBranch(url, branch string) string {
	if branch == "" || !strings.Contains(url, BranchPlaceholder) {
		return url
	}

	// try to normalize branch name for safe use in URLs
	branch = strings.ToLower(strings.TrimSpace(branch))
	branch = strings.ReplaceAll(branch, "_", "-")

	return strings.ReplaceAll(url, BranchPlaceholder, branch)
}
