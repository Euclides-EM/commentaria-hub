package model

type VCSStatus struct {
	Success    bool       `json:"success"`
	BranchName string     `json:"branch"`
	PR         *PRDetails `json:"pr,omitempty"`
}

type PRDetails struct {
	URL    string `json:"url"`
	Number int    `json:"number"`
}
