package filesys

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"github.com/stretchr/testify/require"
)

const transcriptionALTO = `<?xml version="1.0" encoding="UTF-8"?>
<alto xmlns="http://www.loc.gov/standards/alto/ns-v3#">
  <Layout>
    <Page ID="p1" WIDTH="100" HEIGHT="200">
      <PrintSpace>
        <TextBlock ID="b1">
          <TextLine ID="l1"><String CONTENT="Euclid"/></TextLine>
        </TextBlock>
      </PrintSpace>
    </Page>
  </Layout>
</alto>`

func TestRetrieveEditionAltoPageFromTranscriptions(t *testing.T) {
	dataDir := t.TempDir()
	m := NewFileSystemManager(dataDir, "", "", "")
	filePath := filepath.Join(dataDir, "transcriptions", "Paris_1536", "page-0007", "original.xml")
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte(transcriptionALTO), 0o644))

	a, gotPath, err := m.RetrieveEditionAltoPage(&model.Edition{Key: "Paris_1536"}, 7)
	require.NoError(t, err)
	require.Equal(t, filePath, gotPath)
	require.Len(t, a.Layout.Page, 1)
	require.Equal(t, "Euclid", a.Layout.Page[0].PrintSpace.TextBlocks[0].Lines[0].Strings[0].Content)
}

func TestRetrieveAnnotationAltoPageFromTranscriptions(t *testing.T) {
	dataDir := t.TempDir()
	m := NewFileSystemManager(dataDir, "", "", "")
	ann := &annotation.Annotation{
		Meta:      common.NewMeta("ocr"),
		DatasetID: "dataset",
	}
	filePath := filepath.Join(dataDir, "dataset", "annotations", "ocr", "transcriptions", "page-0007", "original.xml")
	require.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0o755))
	require.NoError(t, os.WriteFile(filePath, []byte(transcriptionALTO), 0o644))

	a, gotPath, err := m.RetrieveAnnotationTranscriptionAltoPage(ann, "7")
	require.NoError(t, err)
	require.Equal(t, filePath, gotPath)
	require.Equal(t, "Euclid", a.Layout.Page[0].PrintSpace.TextBlocks[0].Lines[0].Strings[0].Content)
}
