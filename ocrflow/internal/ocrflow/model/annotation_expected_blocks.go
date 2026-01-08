package model

type AnnotationExpectedBlocks struct {
	Category       string                               `json:"category"`
	SanityChecks   []AnnotationExpectedBlocksSanityType `json:"sanity_checks"`
	ExpectedBlocks [][]string                           `json:"expected_blocks"`
	SuggestedDiffs []*SuggestedDiff                     `json:"suggested_diffs,omitempty"`
	FailedChecks   []AnnotationExpectedBlocksSanityType `json:"failed_checks,omitempty"`
}

type SuggestedDiff struct {
	Page        int      `json:"page"`
	TextBlockID string   `json:"text_block_id"`
	Old         []string `json:"old"`
	Correction  []string `json:"correction"`
}

type AnnotationExpectedBlocksSanityType string

const (
	AnnotationExpectedBlocksSanityTypeExact AnnotationExpectedBlocksSanityType = "exact_block_count"
)
