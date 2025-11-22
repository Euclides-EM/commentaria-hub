package ghwrapper

import (
	"fmt"
	"net/url"
	"strings"
)

// IsGitHubTreeURL returns true if the given URL is a valid GitHub tree URL.
// URL must look like: https://www.github.com/<owner>/<repo>/tree/<ref>/<path>
func IsGitHubTreeURL(raw string) bool {
	_, err := parseGitHubTreeURL(raw)
	return err == nil
}

type ghURLTree struct {
	owner, repo, ref, path string
}

func parseGitHubTreeURL(raw string) (*ghURLTree, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}
	if !strings.EqualFold(u.Host, githubHostname) {
		return nil, fmt.Errorf("not a %s url", githubHostname)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 5 || parts[2] != "tree" {
		return nil, fmt.Errorf("URL must look like https://%s/<owner>/<repo>/tree/<ref>/<path>", githubHostname)
	}
	return &ghURLTree{
		owner: parts[0],
		repo:  parts[1],
		ref:   parts[3],
		path:  strings.Join(parts[4:], "/"),
	}, nil
}
