package ghwrapper

import "testing"

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
