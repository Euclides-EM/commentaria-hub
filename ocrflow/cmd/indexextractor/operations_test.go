package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type sequenceExecutor struct {
	responses   []string
	calls       int
	prompts     []string
	attachments []string
	providers   []string
	models      []string
}

func (e *sequenceExecutor) Exec(provider, model, prompt, attachment string) (string, error) {
	if e.calls >= len(e.responses) {
		return "", errors.New("unexpected exec call")
	}
	e.prompts = append(e.prompts, prompt)
	e.attachments = append(e.attachments, attachment)
	e.providers = append(e.providers, provider)
	e.models = append(e.models, model)
	response := e.responses[e.calls]
	e.calls++
	return response, nil
}

type failSecondExecutor struct{ calls int }

func (e *failSecondExecutor) Exec(_, _, _, _ string) (string, error) {
	e.calls++
	if e.calls == 1 {
		return "Durand, 42", nil
	}
	return "", errors.New("upstream timeout")
}

type scriptedResult struct {
	response string
	err      error
}
type scriptedExecutor struct {
	results     []scriptedResult
	calls       int
	attachments []string
}

func (e *scriptedExecutor) Exec(_, _, _ string, attachment string) (string, error) {
	e.attachments = append(e.attachments, attachment)
	result := e.results[e.calls]
	e.calls++
	return result.response, result.err
}

func TestExtractIndexEntriesTwoPassUsesTranscriptionWithoutSecondAttachment(t *testing.T) {
	client := &sequenceExecutor{responses: []string{
		"Durand, **42**",
		`{"entries":[{"name":"Durand","page_number":"42","reference":"","is_bold":true}]}`,
	}}

	entries, issues, transcription, err := extractIndexEntriesWithAudit(
		config{provider: "test", model: "test", extractionMode: modeTwoPass},
		client,
		imageInput{Path: "page.jpg", Volume: "vol_1"},
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("extractIndexEntriesWithAudit returned error: %v", err)
	}
	if len(entries) != 1 || len(issues) != 0 || transcription != "Durand, **42**" {
		t.Fatalf("unexpected result: entries=%#v issues=%#v transcription=%q", entries, issues, transcription)
	}
	if client.calls != 2 || client.attachments[0] != "page.jpg" || client.attachments[1] != "" {
		t.Fatalf("unexpected calls or attachments: calls=%d attachments=%#v", client.calls, client.attachments)
	}
	if client.prompts[0] != transcriptionPrompt || !strings.Contains(client.prompts[1], transcription) {
		t.Fatalf("unexpected prompts: %#v", client.prompts)
	}
}

func TestTwoPassCheckpointsAndReusesTranscriptionAfterSecondPassFailure(t *testing.T) {
	client := &failSecondExecutor{}
	checkpointed := ""
	_, _, _, err := extractIndexEntriesWithCheckpoint(
		config{provider: "ollama", model: "vision", extractionMode: modeTwoPass}, client,
		imageInput{Path: "page.jpg", Volume: "vol_1"}, "", &bytes.Buffer{},
		func(value string) error { checkpointed = value; return nil },
	)
	if err == nil || checkpointed != "Durand, 42" {
		t.Fatalf("expected second-pass failure after checkpoint, err=%v checkpoint=%q", err, checkpointed)
	}

	resumed := &sequenceExecutor{responses: []string{`{"entries":[{"name":"Durand","page_number":"42","reference":"","is_bold":false}]}`}}
	entries, _, _, err := extractIndexEntriesWithCheckpoint(
		config{provider: "ollama", model: "vision", secondProvider: "openai", secondModel: "gpt-test", extractionMode: modeTwoPass},
		resumed, imageInput{Path: "page.jpg", Volume: "vol_1"}, checkpointed, &bytes.Buffer{}, nil,
	)
	if err != nil || len(entries) != 1 || resumed.calls != 1 || resumed.attachments[0] != "" {
		t.Fatalf("resume did not start at pass two: entries=%#v calls=%d attachments=%#v err=%v", entries, resumed.calls, resumed.attachments, err)
	}
	if resumed.providers[0] != "openai" || resumed.models[0] != "gpt-test" {
		t.Fatalf("second pass used %s/%s", resumed.providers[0], resumed.models[0])
	}
}

func TestTwoPassCanUseDifferentProvidersAndModels(t *testing.T) {
	client := &sequenceExecutor{responses: []string{"Durand, 42", `{"entries":[]}`}}
	cfg := config{provider: "ollama", model: "default", firstProvider: "ollama", firstModel: "vision", secondProvider: "openai", secondModel: "accurate", extractionMode: modeTwoPass}
	if _, _, _, err := extractIndexEntriesWithAudit(cfg, client, imageInput{Path: "page.jpg", Volume: "vol_1"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(client.providers, ",") != "ollama,openai" || strings.Join(client.models, ",") != "vision,accurate" {
		t.Fatalf("unexpected provider/model calls: %#v %#v", client.providers, client.models)
	}
}

func TestPassSpecificSettingsOverrideBaseIndependently(t *testing.T) {
	cfg := config{provider: "openai", model: "base-model", firstProvider: "ollama", secondModel: "structured-model", extractionMode: modeTwoPass}
	if firstPassProvider(cfg) != "ollama" || firstPassModel(cfg) != "base-model" || secondPassProvider(cfg) != "openai" || secondPassModel(cfg) != "structured-model" {
		t.Fatalf("unexpected effective settings: first=%s/%s second=%s/%s", firstPassProvider(cfg), firstPassModel(cfg), secondPassProvider(cfg), secondPassModel(cfg))
	}
}

func TestParseCLIRejectsUnknownExtractionMode(t *testing.T) {
	cfg, err := parseCLI([]string{"extract", "--extraction-mode", "three-pass"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parseCLI returned error before config validation: %v", err)
	}
	if err := validateConfig(cfg); err == nil || !strings.Contains(err.Error(), "unsupported extraction mode") {
		t.Fatalf("validateConfig returned %v", err)
	}
}

func TestExtractIndexEntriesSkipsAndReportsInvalidEntries(t *testing.T) {
	client := &sequenceExecutor{responses: []string{
		`{"entries":[{"name":"Broken","page_number":"1","reference":"Other","is_bold":false},{"name":"Valid","page_number":"2","reference":"","is_bold":false}]}`,
	}}
	var out bytes.Buffer

	entries, issues, err := extractIndexEntriesWithIssues(config{provider: "test", model: "test"}, client, imageInput{Path: "page.jpg", Volume: "vol_1"}, &out)
	if err != nil {
		t.Fatalf("extractIndexEntriesWithIssues returned error: %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("got %d calls, want 1", client.calls)
	}
	if len(entries) != 1 || entries[0].Name != "Valid" || entries[0].Volume != "vol_1" {
		t.Fatalf("unexpected entries: %#v", entries)
	}
	if len(issues) != 1 || !strings.Contains(issues[0], `entry 1 requires name`) {
		t.Fatalf("unexpected issues: %#v", issues)
	}
}

func TestExtractIndexEntriesRetriesStructurallyInvalidResponse(t *testing.T) {
	client := &sequenceExecutor{responses: []string{
		`not json`,
		`still not json`,
	}}

	_, err := extractIndexEntries(config{}, client, imageInput{Path: "page.jpg", Volume: "vol_1"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("extractIndexEntries returned nil error")
	}
	if !strings.Contains(err.Error(), `after 2 attempts`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseLettersResponseSkipsAndReportsIncompleteEntries(t *testing.T) {
	raw := `{"entries":[{"letter_number":"18","letter_name":"Valid","page_number":"42"},{"letter_number":"19","letter_name":"","page_number":"43"}]}`

	entries, issues, err := parseLettersResponseWithIssues(raw, "vol_10")
	if err != nil {
		t.Fatalf("parseLettersResponseWithIssues returned error: %v", err)
	}
	if len(entries) != 1 || entries[0].LetterNumber != "18" || entries[0].Volume != "vol_10" {
		t.Fatalf("unexpected entries: %#v", entries)
	}
	if len(issues) != 1 || !strings.Contains(issues[0], `entry 2 requires letter_number`) || !strings.Contains(issues[0], `letter_name=""`) {
		t.Fatalf("unexpected issues: %#v", issues)
	}
}

func TestReportIssuesIncludesCountImageAndDetail(t *testing.T) {
	var out bytes.Buffer
	reportIssues("index", []pageIssues{
		{ImagePath: "vol_10/page.jpg", Issues: []string{"entry 19 is incomplete", "entry 20 is invalid"}},
	}, &out)

	output := out.String()
	if !strings.Contains(output, "index: 2 tolerated parsing issues across 1 affected images") ||
		!strings.Contains(output, "vol_10/page.jpg: entry 19 is incomplete") {
		t.Fatalf("unexpected issue report: %q", output)
	}
}

func TestIndexExtractionRecordsFailureAndContinues(t *testing.T) {
	dir := t.TempDir()
	raw := filepath.Join(dir, "raw", "vol_1")
	if err := os.MkdirAll(raw, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"01.jpg", "02.jpg"} {
		if err := os.WriteFile(filepath.Join(raw, name), []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config{kind: kindIndex, indexDir: filepath.Join(dir, "raw"), indexCSV: filepath.Join(dir, "index.csv"), provider: "ollama", model: "qwen", extractionMode: modeTwoPass, resume: true}
	client := &scriptedExecutor{results: []scriptedResult{
		{err: errors.New("502 timeout")},
		{response: "Durand, 42"},
		{response: `{"entries":[{"name":"Durand","page_number":"42","reference":"","is_bold":false}]}`},
	}}
	if err := runIndexExtraction(cfg, client, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	manifest, err := loadIndexManifest(cfg.indexCSV, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Pages) != 2 || manifest.Pages[0].Failure == nil || manifest.Pages[0].Failure.Phase != failurePhaseFirst || manifest.Pages[1].Entries == nil {
		t.Fatalf("unexpected manifest pages: %#v", manifest.Pages)
	}
	var status bytes.Buffer
	if err := reportStatus(cfg, &status); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(status.String(), "1 failed images") || !strings.Contains(status.String(), "first-pass") {
		t.Fatalf("unexpected failures: %s", status.String())
	}
	var validation bytes.Buffer
	if err := validateOutputs(cfg, &validation); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(validation.String(), "1 failed images") || !strings.Contains(validation.String(), "first-pass") {
		t.Fatalf("validation omitted failures: %s", validation.String())
	}
	cfg.skipFailures = true
	skipped := &scriptedExecutor{}
	if err := runIndexExtraction(cfg, skipped, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if skipped.calls != 0 {
		t.Fatalf("--skip-failures made %d unexpected calls", skipped.calls)
	}
}

func TestSecondPassFailureResumesWithoutFirstPass(t *testing.T) {
	dir := t.TempDir()
	raw := filepath.Join(dir, "raw", "vol_1")
	if err := os.MkdirAll(raw, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raw, "page.jpg"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config{kind: kindIndex, indexDir: filepath.Join(dir, "raw"), indexCSV: filepath.Join(dir, "index.csv"), provider: "ollama", model: "qwen", extractionMode: modeTwoPass, resume: true}
	first := &scriptedExecutor{results: []scriptedResult{{response: "Durand, 42"}, {err: errors.New("502 timeout")}}}
	if err := runIndexExtraction(cfg, first, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	manifest, err := loadIndexManifest(cfg.indexCSV, true)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Pages[0].Failure == nil || manifest.Pages[0].Failure.Phase != failurePhaseSecond || manifest.Pages[0].Transcription == "" {
		t.Fatalf("unexpected failed page: %#v", manifest.Pages[0])
	}
	second := &scriptedExecutor{results: []scriptedResult{{response: `{"entries":[]}`}}}
	if err := runIndexExtraction(cfg, second, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if second.calls != 1 || second.attachments[0] != "" {
		t.Fatalf("retry repeated first pass: calls=%d attachments=%#v", second.calls, second.attachments)
	}
}
