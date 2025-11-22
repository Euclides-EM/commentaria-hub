package models

type Dataset struct {
	ID         string `json:"id"`
	LocalPath  string `json:"local_path"`
	RemotePath string `json:"remote_path"`
}
