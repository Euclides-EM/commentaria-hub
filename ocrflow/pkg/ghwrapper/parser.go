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
	switch u.Hostname() {
	case "raw.githubusercontent.com":
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) < 3 {
			return nil, fmt.Errorf("URL must look like https://raw.githubusercontent.com/<owner>/<repo>/<ref>/<path>")
		}
		return &ghURLTree{
			owner: parts[0],
			repo:  parts[1],
			ref:   parts[2],
			path:  strings.Join(parts[3:], "/"),
		}, nil
	case "api.github.com":
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) < 5 || parts[0] != "repos" || parts[3] != "contents" {
			return nil, fmt.Errorf("URL must look like https://api.github.com/repos/<owner>/<repo>/contents/<path>?ref=<ref>")
		}
		ref := u.Query().Get("ref")
		return &ghURLTree{
			owner: parts[1],
			repo:  parts[2],
			ref:   ref,
			path:  strings.Join(parts[4:], "/"),
		}, nil
	case "github.com":
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) < 5 || (parts[2] != "tree" && parts[2] != "blob") {
			return nil, fmt.Errorf("URL must look like https://github.com/<owner>/<repo>/<tree|blob>/<ref>>/<path>")
		}
		return &ghURLTree{
			owner: parts[0],
			repo:  parts[1],
			ref:   parts[3],
			path:  strings.Join(parts[4:], "/"),
		}, nil
	}
	return nil, fmt.Errorf("invalid hostname: %s", u.Hostname())
}
