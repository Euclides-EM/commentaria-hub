package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsingValidationIsTextOnlyAndProducesDifferences(t *testing.T) {
	client := &sequenceExecutor{responses: []string{`{"accurate":false,"summary":"one omission","differences":[{"location":"Durand","expected":"page 42","actual":"<missing>","reason":"entry omitted"}]}`}}
	transcription := "Durand, 42"
	structured := `[]`
	digest := validationSourceDigest(validationParsing, transcription, structured)
	result := executeAIValidation(config{provider: "openai", model: "gpt-test"}, client, validationParsing, imageInput{Path: "page.jpg", Volume: "vol_1"}, transcription, structured, digest)

	if result.Failure != "" || result.Accurate || len(result.Differences) != 1 {
		t.Fatalf("unexpected validation result: %#v", result)
	}
	if client.attachments[0] != "" || !strings.Contains(client.prompts[0], transcription) || !strings.Contains(client.prompts[0], structured) {
		t.Fatalf("parsing validation was not text-only: attachments=%#v prompt=%q", client.attachments, client.prompts[0])
	}
}

func TestTranscriptionValidationAttachesImage(t *testing.T) {
	client := &sequenceExecutor{responses: []string{`{"accurate":true,"summary":"matches","differences":[]}`}}
	result := executeAIValidation(config{provider: "ollama", model: "vision"}, client, validationTranscription, imageInput{Path: "page.jpg", Volume: "vol_1"}, "Durand, 42", "ignored", "digest")
	if result.Failure != "" || !result.Accurate || client.attachments[0] != "page.jpg" {
		t.Fatalf("unexpected result or attachment: result=%#v attachments=%#v", result, client.attachments)
	}
}

func TestValidationFailurePersistsRawResponse(t *testing.T) {
	raw := `{"accurate":false,"summary":"bad shape"}`
	client := &sequenceExecutor{responses: []string{raw}}
	result := executeAIValidation(config{provider: "ollama", model: "vision"}, client, validationTranscription, imageInput{Path: "page.jpg", Volume: "vol_1"}, "Durand, 42", "", "digest")

	if result.Failure != "validation response requires a differences array" || result.RawResponse != raw {
		t.Fatalf("validation failure did not retain raw response: %#v", result)
	}

	dir := t.TempDir()
	output := filepath.Join(dir, "index.csv")
	state := validationManifest{Version: validationManifestVersion, Kind: kindIndex, Check: validationTranscription, Results: []validationResult{result}}
	if err := saveValidationCheckpoint(output, state); err != nil {
		t.Fatal(err)
	}
	checkpoint, err := os.ReadFile(validationStatePath(output, validationTranscription))
	if err != nil {
		t.Fatal(err)
	}
	review, err := os.ReadFile(validationReviewPath(output, validationTranscription))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(checkpoint, []byte(`"raw_response"`)) || !bytes.Contains(review, []byte("### Raw response")) || !bytes.Contains(review, []byte(raw)) {
		t.Fatalf("raw response missing from checkpoint or review:\ncheckpoint=%s\nreview=%s", checkpoint, review)
	}
}

func TestValidationCheckpointWritesHumanReviewAndDigestControlsResume(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "index.csv")
	state := validationManifest{Version: validationManifestVersion, Kind: kindIndex, Check: validationParsing, Results: []validationResult{{
		ImagePath: "vol_1/page.jpg", SourceDigest: "same", Accurate: false, Summary: "bad", Differences: []validationDifference{{Location: "entry 1", Expected: "42", Actual: "24", Reason: "reversed"}},
	}}}
	if err := saveValidationCheckpoint(output, state); err != nil {
		t.Fatal(err)
	}
	review, err := os.ReadFile(validationReviewPath(output, validationParsing))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(review, []byte("REVIEW REQUIRED")) || !bytes.Contains(review, []byte("Expected: `42`")) {
		t.Fatalf("review is not human-readable:\n%s", review)
	}
	cfg := config{resume: true}
	if !shouldSkipValidation(state, "vol_1/page.jpg", "same", cfg, false) {
		t.Fatal("unchanged completed validation should resume")
	}
	if shouldSkipValidation(state, "vol_1/page.jpg", "changed", cfg, false) {
		t.Fatal("changed source must be revalidated")
	}
}

func TestParseCLIAcceptsAIValidationCommands(t *testing.T) {
	for _, command := range []string{commandValidateTranscriptions, commandValidateParsing} {
		cfg, err := parseCLI([]string{command, "--kind", "index"}, &bytes.Buffer{})
		if err != nil || cfg.command != command {
			t.Fatalf("command %s: cfg=%#v err=%v", command, cfg, err)
		}
	}
}

func TestDefaultValidationArtifactsUseReviewsDirectory(t *testing.T) {
	output := filepath.Join("data", "outputs", "index.csv")
	want := filepath.Join("data", "reviews", "index.csv.validate-parsing.json")
	if got := validationStatePath(output, validationParsing); got != want {
		t.Fatalf("validationStatePath() = %q, want %q", got, want)
	}
}
