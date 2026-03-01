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
