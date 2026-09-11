package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	"github.com/stretchr/testify/require"
)

const alternativeTranscriptionALTO = `<?xml version="1.0" encoding="UTF-8"?>
<alto xmlns="http://www.loc.gov/standards/alto/ns-v3#">
  <Layout>
    <Page ID="p1" WIDTH="100" HEIGHT="200">
      <PrintSpace>
        <TextBlock ID="b1">
          <TextLine ID="l1"><String CONTENT="ALTO alternative"/></TextLine>
        </TextBlock>
      </PrintSpace>
    </Page>
  </Layout>
</alto>`

func TestBuildTEIFromMultipleMixedTranscriptionFiles(t *testing.T) {
	dir := t.TempDir()
	paths := map[string]string{
		"original.md":     "## Markdown original\n\nOriginal body",
		"review.txt":      "TXT alternative",
		"second-pass.xml": alternativeTranscriptionALTO,
	}
	for name, content := range paths {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}

	doc, err := buildTEIFromTranscriptionFiles("32", []filesys.TranscriptionFile{
		{Name: "original.md", Label: "original", Path: filepath.Join(dir, "original.md"), Format: filesys.TranscriptionFileFormatMarkdown},
		{Name: "review.txt", Label: "review", Path: filepath.Join(dir, "review.txt"), Format: filesys.TranscriptionFileFormatText},
		{Name: "second-pass.xml", Label: "second-pass", Path: filepath.Join(dir, "second-pass.xml"), Format: filesys.TranscriptionFileFormatALTO},
	}, nil, "image.png", nil)
	require.NoError(t, err)
	require.Len(t, doc.Text.Body.Divs, 3)
	require.Equal(t, "transcription", doc.Text.Body.Divs[0].Type)
	require.Equal(t, transcriptionAlternativeDivType, doc.Text.Body.Divs[1].Type)
	require.Equal(t, "review", doc.Text.Body.Divs[1].N)
	require.Equal(t, transcriptionAlternativeDivType, doc.Text.Body.Divs[2].Type)
	require.Equal(t, "second-pass", doc.Text.Body.Divs[2].N)
	for _, div := range doc.Text.Body.Divs[1:] {
		for _, block := range div.Abs {
			require.Empty(t, block.Facs)
			for _, line := range block.Lines {
				require.Empty(t, line.Facs)
			}
		}
	}

	xml, err := doc.ToXML()
	require.NoError(t, err)
	require.Contains(t, string(xml), "Markdown original")
	require.Contains(t, string(xml), "TXT alternative")
	require.Contains(t, string(xml), "ALTO alternative")
}
