package llm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildInputPayloadUsesInputImageForImages(t *testing.T) {
	path := filepath.Join(t.TempDir(), "page.jpg")
	if err := os.WriteFile(path, []byte("image bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	attachment := payloadAttachment(t, path)
	if attachment["type"] != "input_image" {
		t.Fatalf("attachment type = %v, want input_image", attachment["type"])
	}
	imageURL, ok := attachment["image_url"].(string)
	if !ok || !strings.HasPrefix(imageURL, "data:image/jpeg;base64,") {
		t.Fatalf("image_url = %v, want JPEG data URL", attachment["image_url"])
	}
	if _, exists := attachment["file_data"]; exists {
		t.Fatal("image attachment unexpectedly contains file_data")
	}
}

func TestBuildInputPayloadUsesInputFileForDocuments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	attachment := payloadAttachment(t, path)
	if attachment["type"] != "input_file" {
		t.Fatalf("attachment type = %v, want input_file", attachment["type"])
	}
	fileData, ok := attachment["file_data"].(string)
	if !ok || !strings.HasPrefix(fileData, "data:text/plain") {
		t.Fatalf("file_data = %v, want text data URL", attachment["file_data"])
	}
	if attachment["filename"] != "notes.txt" {
		t.Fatalf("filename = %v, want notes.txt", attachment["filename"])
	}
}

func payloadAttachment(t *testing.T, path string) map[string]any {
	t.Helper()
	payload, err := buildInputPayload(Prompt{Dynamic: "prompt"}, path, false)
	if err != nil {
		t.Fatal(err)
	}
	messages, ok := payload.([]map[string]any)
	if !ok || len(messages) != 1 {
		t.Fatalf("payload = %#v, want one message", payload)
	}
	content, ok := messages[0]["content"].([]map[string]any)
	if !ok || len(content) != 2 {
		t.Fatalf("content = %#v, want prompt and attachment", messages[0]["content"])
	}
	return content[1]
}

func TestBuildInputPayloadPlacesCacheableStaticPrefixBeforeDynamicImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "page.jpg")
	if err := os.WriteFile(path, []byte("image bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	payload, err := buildInputPayload(Prompt{Static: "stable dialect", Dynamic: "page input"}, path, true)
	if err != nil {
		t.Fatal(err)
	}
	messages := payload.([]map[string]any)
	if len(messages) != 2 || messages[0]["role"] != "developer" || messages[1]["role"] != "user" {
		t.Fatalf("messages = %#v, want developer prefix followed by user input", messages)
	}
	staticContent := messages[0]["content"].([]map[string]any)
	if _, ok := staticContent[0]["prompt_cache_breakpoint"]; !ok {
		t.Fatalf("static block = %#v, want explicit cache breakpoint", staticContent[0])
	}
	dynamicContent := messages[1]["content"].([]map[string]any)
	if dynamicContent[0]["type"] != "input_text" || dynamicContent[1]["type"] != "input_image" {
		t.Fatalf("dynamic content = %#v, want text followed by image", dynamicContent)
	}
}

func TestSupportsExplicitPromptCachingOnlyForGPT56Family(t *testing.T) {
	if !supportsExplicitPromptCaching("gpt-5.6") || !supportsExplicitPromptCaching("gpt-5.6-2026-08-01") {
		t.Fatal("gpt-5.6 family should support explicit prompt caching")
	}
	if supportsExplicitPromptCaching("gpt-5.5") || supportsExplicitPromptCaching("gpt-4.1") {
		t.Fatal("earlier models should use implicit prompt caching")
	}
}
