package ghwrapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func gitExec(repoDir, cmd string, args ...string) (stdout, stderr string, err error) {
	c := exec.Command(cmd, args...)
	c.Dir = repoDir
	var outBuf, errBuf bytes.Buffer
	c.Stdout = &outBuf
	c.Stderr = &errBuf
	err = c.Run()
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), err
}

// GetCurrentBranch returns the current git branch name.
func GetCurrentBranch(repoDir string) (string, error) {
	stdout, _, err := gitExec(repoDir, "git", "branch", "--show-current")
	return stdout, err
}

// GetRepoOwnerRepo returns owner and repo name from remote.origin.url.
func GetRepoOwnerRepo(repoDir string) (owner, repo string, err error) {
	stdout, _, err := gitExec(repoDir, "git", "config", "--get", "remote.origin.url")
	if err != nil {
		return "", "", err
	}
	idx := strings.Index(stdout, "github.com")
	if idx < 0 {
		return "", "", fmt.Errorf("remote.origin.url does not contain github.com: %s", stdout)
	}
	// Match github.com:owner/repo or github.com/owner/repo
	rest := stdout[idx+10:]
	rest = strings.TrimPrefix(rest, "/")
	rest = strings.TrimPrefix(rest, ":")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 2 {
		repo = strings.TrimSuffix(parts[1], ".git")
		return parts[0], repo, nil
	}
	return "", "", fmt.Errorf("remote.origin.url does not match expected format: %s", stdout)
}

// GetExistingPR returns open PR for branch if any.
func GetExistingPR(owner, repo, branch, token string) (number int, htmlURL string, err error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?head=%s:%s&state=open", owner, repo, owner, branch)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}
	var prs []struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return 0, "", err
	}
	if len(prs) > 0 {
		return prs[0].Number, prs[0].HTMLURL, nil
	}
	return 0, "", nil
}

// StatusPorcelain returns output of git status --porcelain for the given paths.
func StatusPorcelain(repoDir string, paths ...string) (string, error) {
	args := append([]string{"status", "--porcelain"}, paths...)
	stdout, _, err := gitExec(repoDir, "git", args...)
	return stdout, err
}

// CreateBranch creates and checks out a new branch.
func CreateBranch(repoDir, name string) error {
	_, _, err := gitExec(repoDir, "git", "checkout", "-b", name)
	return err
}

// PushBranch pushes the current branch to origin (with -u on first push).
func PushBranch(repoDir, branch string, setUpstream bool) error {
	args := []string{"push"}
	if setUpstream {
		args = append(args, "-u", "origin", branch)
	} else {
		args = append(args, "origin", branch)
	}
	_, _, err := gitExec(repoDir, "git", args...)
	return err
}

// AddAndCommit adds paths and commits with message.
func AddAndCommit(repoDir string, paths []string, message string) error {
	if len(paths) == 0 {
		return nil
	}
	args := append([]string{"add"}, paths...)
	_, _, err := gitExec(repoDir, "git", args...)
	if err != nil {
		return err
	}
	_, _, err = gitExec(repoDir, "git", "commit", "-m", message)
	return err
}

// CreatePullRequest creates a PR via GitHub API.
func CreatePullRequest(owner, repo, head, token, title, body string) (number int, htmlURL string, err error) {
	payload := map[string]string{
		"title": title,
		"head":  head,
		"base":  "main",
		"body":  body,
	}
	bodyBytes, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}
	var pr struct {
		Number  int    `json:"number"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return 0, "", err
	}
	return pr.Number, pr.HTMLURL, nil
}

func Checkout(repoDir, branch string) error {
	_, _, err := gitExec(repoDir, "git", "checkout", branch)
	return err
}

func Pull(repoDir string) error {
	_, _, err := gitExec(repoDir, "git", "pull")
	return err
}
