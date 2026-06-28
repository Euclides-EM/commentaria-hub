package model

type Pseudonym struct {
	Name      string `json:"name"`
	Pseudonym string `json:"pseudonym"`
	Position  string `json:"position"`
	Source    string `json:"source"`
}
