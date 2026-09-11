package filesys

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListTranscriptionFilesSupportsMultipleMixedFormats(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{
		"notes.json",
		"z_second.txt",
		"another_option.md",
		"facsimile.xml",
		"original.txt",
		"original.md",
		"original.xml",
	} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("test"), 0o600))
	}

	files, err := listTranscriptionFiles(dir)
	require.NoError(t, err)
	require.Equal(t, []string{
		"original.xml",
		"original.md",
		"original.txt",
		"another_option.md",
		"facsimile.xml",
		"z_second.txt",
	}, transcriptionFileNames(files))
	require.Equal(t, []string{
		"original.xml",
		"original.md",
		"original.txt",
		"another_option",
		"facsimile",
		"z_second",
	}, transcriptionFileLabels(files))
	require.Equal(t, []TranscriptionFileFormat{
		TranscriptionFileFormatALTO,
		TranscriptionFileFormatMarkdown,
		TranscriptionFileFormatText,
		TranscriptionFileFormatMarkdown,
		TranscriptionFileFormatALTO,
		TranscriptionFileFormatText,
	}, transcriptionFileFormats(files))
}

func transcriptionFileNames(files []TranscriptionFile) []string {
	result := make([]string, len(files))
	for i := range files {
		result[i] = files[i].Name
	}
	return result
}

func transcriptionFileLabels(files []TranscriptionFile) []string {
	result := make([]string, len(files))
	for i := range files {
		result[i] = files[i].Label
	}
	return result
}

func transcriptionFileFormats(files []TranscriptionFile) []TranscriptionFileFormat {
	result := make([]TranscriptionFileFormat, len(files))
	for i := range files {
		result[i] = files[i].Format
	}
	return result
}
