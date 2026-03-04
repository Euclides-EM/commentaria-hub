package service

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
)

type VCSMgt struct {
	repoPath              string
	itemsMetadataStoreDir string
	titlePageImgDir       string
}

// NewVCSMgt creates a repo service. itemsMetadataStoreDir may be empty to disable.
func NewVCSMgt(itemsMetadataStoreDir, titlePageImgDir string) *VCSMgt {
	repoPath := futils.SharedParent(itemsMetadataStoreDir, titlePageImgDir)
	relIMSD, err := filepath.Rel(repoPath, itemsMetadataStoreDir)
	if err != nil {
		log.Fatalf("filepath.Rel(%q, %q): %v", repoPath, itemsMetadataStoreDir, err)
	}
	relTPID, err := filepath.Rel(repoPath, titlePageImgDir)
	if err != nil {
		log.Fatalf("filepath.Rel(%q, %q): %v", repoPath, titlePageImgDir, err)
	}
	return &VCSMgt{
		itemsMetadataStoreDir: relIMSD,
		titlePageImgDir:       relTPID,
	}
}

func (r *VCSMgt) GetCommitSHA(repoPath string) (string, error) {
	return ghwrapper.GetLatestCommitSHA(repoPath)
}

// Pull runs git pull and returns the branch name (after possibly checking out main).
func (r *VCSMgt) Pull(token string) (*model.VCSStatus, error) {
	branch, err := ghwrapper.GetCurrentBranch(r.repoPath)
	if err != nil {
		return nil, fmt.Errorf("get current branch: %w", err)
	}
	var prNum int
	var prURL string
	if branch != "main" {
		owner, repo, err := ghwrapper.GetRepoOwnerRepo(r.repoPath)
		if err != nil {
			return nil, fmt.Errorf("get repo owner/repo: %w", err)
		}
		prNum, prURL, _ = ghwrapper.GetExistingPR(owner, repo, branch, token)
		if prNum == 0 {
			err = ghwrapper.Checkout(r.repoPath, "main")
			if err != nil {
				return nil, fmt.Errorf("checkout main: %w", err)
			}
			branch = "main"
		}
	}
	var pr *model.PRDetails
	if prNum != 0 {
		pr = &model.PRDetails{
			Number: prNum,
			URL:    prURL,
		}
	}
	err = ghwrapper.Pull(r.repoPath)
	if err != nil {
		return nil, fmt.Errorf("git pull: %w", err)
	}
	return &model.VCSStatus{
		Success:    true,
		BranchName: branch,
		PR:         pr,
	}, nil
}

func (r *VCSMgt) Push(token string) (*model.VCSStatus, error) {
	status, err := r.Pull(token)
	if err != nil {
		return nil, fmt.Errorf("repo pull: %w", err)
	}
	owner, repo, err := ghwrapper.GetRepoOwnerRepo(r.repoPath)
	if err != nil {
		return nil, err
	}
	statusOut, err := ghwrapper.StatusPorcelain(r.repoPath, r.itemsMetadataStoreDir, r.titlePageImgDir)
	if err != nil {
		return nil, fmt.Errorf("git status: %w", err)
	}
	if strings.TrimSpace(statusOut) == "" {
		return status, nil
	}
	if status.BranchName == "main" {
		// editor-YYYYMMDD-HHMM format (from TS: slice(0,16).replace(/[-:]/g,"").replace("T","-"))
		ts := time.Now().UTC().Format("2006-01-02T15:04")
		ts = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(ts, "-", ""), ":", ""), "T", "-")
		status.BranchName = "editor-" + ts
		if err := ghwrapper.CreateBranch(r.repoPath, status.BranchName); err != nil {
			return nil, fmt.Errorf("create branch: %w", err)
		}
		if err := ghwrapper.PushBranch(r.repoPath, status.BranchName, true); err != nil {
			return nil, fmt.Errorf("push branch: %w", err)
		}
	}
	if err := ghwrapper.AddAndCommit(r.repoPath, []string{r.titlePageImgDir, r.itemsMetadataStoreDir}, "Update documentation files"); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	if err := ghwrapper.PushBranch(r.repoPath, status.BranchName, false); err != nil {
		return nil, fmt.Errorf("push: %w", err)
	}
	prNum, prURL, _ := ghwrapper.GetExistingPR(owner, repo, status.BranchName, token)
	if prNum == 0 {
		prNum, prURL, err = ghwrapper.CreatePullRequest(owner, repo, status.BranchName, token,
			"Metadata updates - "+status.BranchName,
			"Automated PR for documentation updates created from editor")
		if err != nil {
			return nil, fmt.Errorf("create PR: %w", err)
		}
	}
	status.PR = &model.PRDetails{
		Number: prNum,
		URL:    prURL,
	}
	return status, nil

}
