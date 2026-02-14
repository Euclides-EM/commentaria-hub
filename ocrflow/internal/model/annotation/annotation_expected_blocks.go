package annotation

type ExpectedBlocks struct {
	Category       string                     `json:"category"`
	SanityChecks   []ExpectedBlocksSanityType `json:"sanity_checks"`
	ExpectedBlocks [][]string                 `json:"expected_blocks"`
	SuggestedDiffs []*SuggestedDiff           `json:"suggested_diffs,omitempty"`
	FailedChecks   []ExpectedBlocksSanityType `json:"failed_checks,omitempty"`
}

type SuggestedDiff struct {
	Page        int      `json:"page"`
	TextBlockID string   `json:"text_block_id"`
	Old         []string `json:"old"`
	Correction  []string `json:"correction"`
}

type ExpectedBlocksSanityType string

const (
	ExpectedBlocksSanityTypeExact ExpectedBlocksSanityType = "exact_block_count"
)
