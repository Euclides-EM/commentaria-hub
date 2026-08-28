package futils

import "testing"

func TestIsLocalFileURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"valid file URL", "file:///path/to/file.txt", true},
		{"valid file URL with localhost", "file://localhost/path/to/file.txt", true},
		{"invalid scheme", "http://example.com/file.txt", false},
		{"invalid host", "file://remotehost/path/to/file.txt", false},
		{"invalid URL format", "not a url", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLocalFileURL(tt.url); got != tt.want {
				t.Errorf("IsLocalFileURL(%v) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestLocalFilePathToURL(t *testing.T) {
	got, err := LocalFilePathToURL("/tmp/a file.pdf")
	if err != nil {
		t.Fatalf("LocalFilePathToURL returned error: %v", err)
	}
	want := "file:///tmp/a%20file.pdf"
	if got != want {
		t.Fatalf("LocalFilePathToURL() = %q, want %q", got, want)
	}
}

func TestURLToLocalFilePathUnescapesPath(t *testing.T) {
	got, err := URLToLocalFilePath("file:///tmp/London_1680%E2%80%9381/W%C3%BCrzburg_1661.pdf")
	if err != nil {
		t.Fatalf("URLToLocalFilePath returned error: %v", err)
	}
	want := "/tmp/London_1680–81/Würzburg_1661.pdf"
	if got != want {
		t.Fatalf("URLToLocalFilePath() = %q, want %q", got, want)
	}
}
