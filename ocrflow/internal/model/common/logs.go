package common

type LogTail struct {
	Count   int      `json:"count"`
	Lines   []string `json:"lines"`
	Service string   `json:"service"`
}
