package app

import "time"

const manifestVersion = 5

const legacyManifestVersion = 1
const twoPassManifestVersion = 2
const resumableManifestVersion = 3
const failureManifestVersion = 4

const responseValidationAttempts = 2

var (
	indexHeader   = []string{"name", "page_number", "reference", "is_bold", "volume"}
	lettersHeader = []string{"letter_number", "letter_name", "page_number", "volume"}
)

type indexManifest struct {
	Version int               `json:"version"`
	Kind    string            `json:"kind"`
	Pages   []indexPageResult `json:"pages"`
}

type indexPageResult struct {
	ImagePath             string       `json:"image_path"`
	Volume                string       `json:"volume"`
	Provider              string       `json:"provider"`
	Model                 string       `json:"model"`
	ExtractedAt           time.Time    `json:"extracted_at"`
	ExtractionMode        string       `json:"extraction_mode"`
	Transcription         string       `json:"transcription,omitempty"`
	TranscriptionProvider string       `json:"transcription_provider,omitempty"`
	TranscriptionModel    string       `json:"transcription_model,omitempty"`
	TranscribedAt         time.Time    `json:"transcribed_at,omitempty"`
	Entries               []indexEntry `json:"entries"`
	Issues                []string     `json:"issues,omitempty"`
	Failure               *pageFailure `json:"failure,omitempty"`
}

type lettersManifest struct {
	Version int                 `json:"version"`
	Kind    string              `json:"kind"`
	Pages   []lettersPageResult `json:"pages"`
}

type lettersPageResult struct {
	ImagePath             string        `json:"image_path"`
	Volume                string        `json:"volume"`
	Provider              string        `json:"provider"`
	Model                 string        `json:"model"`
	ExtractedAt           time.Time     `json:"extracted_at"`
	ExtractionMode        string        `json:"extraction_mode"`
	Transcription         string        `json:"transcription,omitempty"`
	TranscriptionProvider string        `json:"transcription_provider,omitempty"`
	TranscriptionModel    string        `json:"transcription_model,omitempty"`
	TranscribedAt         time.Time     `json:"transcribed_at,omitempty"`
	Entries               []letterEntry `json:"entries"`
	Issues                []string      `json:"issues,omitempty"`
	Failure               *pageFailure  `json:"failure,omitempty"`
}

type pageFailure struct {
	Phase    string    `json:"phase"`
	Provider string    `json:"provider"`
	Model    string    `json:"model"`
	Error    string    `json:"error"`
	FailedAt time.Time `json:"failed_at"`
}

type manualOverride struct {
	CorrectedBy string                       `json:"corrected_by"`
	CorrectedAt time.Time                    `json:"corrected_at"`
	Reason      string                       `json:"reason,omitempty"`
	Changes     map[string]manualFieldChange `json:"changes"`
}

type manualFieldChange struct {
	Old string `json:"old"`
	New string `json:"new"`
}

const (
	failurePhaseSingle = "single-pass"
	failurePhaseFirst  = "first-pass"
	failurePhaseSecond = "second-pass"
)

type checkpointError struct{ err error }

func (e checkpointError) Error() string { return e.err.Error() }
func (e checkpointError) Unwrap() error { return e.err }

type pageIssues struct {
	ImagePath string
	Issues    []string
}

type failedPage struct {
	ImagePath string
	Failure   *pageFailure
}
