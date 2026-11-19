package model

type ReprintDetection struct {
	Candidates []ReprintRelationship `json:"candidates" readonly:"true"`
}

type ApplyReprints struct {
	Relationships []ReprintRelationship `json:"relationships"`
	Updated       []string              `json:"updated"`
	Skipped       []string              `json:"skipped"`
}

type ReprintRelationship struct {
	EditionKey string `json:"editionKey"`
	ReprintOf  string `json:"reprintOf"`
}
