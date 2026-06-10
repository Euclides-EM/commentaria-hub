package ghwrapper

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runGit(t *testing.T, repoDir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestParseGitRemoteOwnerRepo(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		wantOwner string
		wantRepo  string
	}{
		{
			name:      "https",
			remoteURL: "https://github.com/Euclides-EM/commentaria-hub.git",
			wantOwner: "Euclides-EM",
			wantRepo:  "commentaria-hub",
		},
		{
			name:      "standard ssh",
			remoteURL: "git@github.com:Euclides-EM/commentaria-hub.git",
			wantOwner: "Euclides-EM",
			wantRepo:  "commentaria-hub",
		},
		{
			name:      "ssh host alias",
			remoteURL: "git@github-commentaria:Euclides-EM/commentaria-hub.git",
			wantOwner: "Euclides-EM",
			wantRepo:  "commentaria-hub",
		},
		{
			name:      "ssh url with host alias",
			remoteURL: "ssh://git@github-commentaria/Euclides-EM/commentaria-hub.git",
			wantOwner: "Euclides-EM",
			wantRepo:  "commentaria-hub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseGitRemoteOwnerRepo(tt.remoteURL)
			if err != nil {
				t.Fatalf("parseGitRemoteOwnerRepo() error = %v", err)
			}
			if owner != tt.wantOwner || repo != tt.wantRepo {
				t.Fatalf("parseGitRemoteOwnerRepo() = %q, %q; want %q, %q", owner, repo, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}

func TestParseGitRemoteOwnerRepoInvalid(t *testing.T) {
	_, _, err := parseGitRemoteOwnerRepo("git@github-commentaria:Euclides-EM")
	if err == nil {
		t.Fatal("parseGitRemoteOwnerRepo() error = nil; want error")
	}
}

func TestPushBranchRejectsMain(t *testing.T) {
	err := PushBranch(t.TempDir(), "main", false)
	if err == nil {
		t.Fatal("PushBranch() error = nil; want error")
	}
}

func TestAddAndCommitSkipsIgnoredNestedDirectories(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@example.com")
	runGit(t, repoDir, "config", "user.name", "Test User")

	if err := os.WriteFile(filepath.Join(repoDir, ".gitignore"), []byte("**/imgs/_variants/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	imgDir := filepath.Join(repoDir, "ocrflow", "store", "data", "tps", "imgs")
	variantDir := filepath.Join(imgDir, "_variants", "thumb")
	if err := os.MkdirAll(variantDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imgDir, "Paris_1622_tp.jpeg"), []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(variantDir, "Paris_1622_tp.jpeg"), []byte("variant"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repoDir, "add", ".gitignore")
	runGit(t, repoDir, "commit", "-m", "Initial commit")

	paths := []string{
		"ocrflow/store/data/tps/imgs",
		":(exclude)ocrflow/store/data/tps/imgs/_variants",
	}
	if err := AddAndCommit(repoDir, paths, "Add title page image"); err != nil {
		t.Fatalf("AddAndCommit() error = %v", err)
	}

	tracked := runGit(t, repoDir, "ls-files")
	if !strings.Contains(tracked, "ocrflow/store/data/tps/imgs/Paris_1622_tp.jpeg") {
		t.Fatalf("tracked files missing title page image:\n%s", tracked)
	}
	if strings.Contains(tracked, "_variants") {
		t.Fatalf("tracked files include ignored variant:\n%s", tracked)
	}
}
