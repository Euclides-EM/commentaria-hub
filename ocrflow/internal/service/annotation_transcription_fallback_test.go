package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTranscriptionFallbackFormatInDir(t *testing.T) {
	dir := t.TempDir()

	format, err := transcriptionFallbackFormatInDir(filepath.Join(dir, "missing"))
	require.NoError(t, err)
	require.Empty(t, format)

	pageDir := filepath.Join(dir, "page-0001")
	require.NoError(t, os.MkdirAll(pageDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pageDir, "original.txt"), []byte("text"), 0o644))

	format, err = transcriptionFallbackFormatInDir(dir)
	require.NoError(t, err)
	require.Equal(t, "text", format)

	require.NoError(t, os.WriteFile(filepath.Join(pageDir, "original.md"), []byte("markdown"), 0o644))

	format, err = transcriptionFallbackFormatInDir(dir)
	require.NoError(t, err)
	require.Equal(t, "markdown", format)

	require.NoError(t, os.WriteFile(filepath.Join(pageDir, "original.xml"), []byte("<alto/>"), 0o644))

	format, err = transcriptionFallbackFormatInDir(dir)
	require.NoError(t, err)
	require.Equal(t, "alto", format)
}
