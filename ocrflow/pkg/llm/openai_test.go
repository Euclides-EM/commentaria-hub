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
	payload, err := buildInputPayload("prompt", path)
	if err != nil {
		t.Fatal(err)
	}
	messages, ok := payload.([]map[string]any)
	if !ok || len(messages) != 1 {
		t.Fatalf("payload = %#v, want one message", payload)
	}
	content, ok := messages[0]["content"].([]map[string]any)
	if !ok || len(content) != 2 {
		t.Fatalf("content = %#v, want attachment and prompt", messages[0]["content"])
	}
	return content[0]
}
