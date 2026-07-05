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

func TestCleanupLocalStorePreservesEditionTranscriptions(t *testing.T) {
	workDir := t.TempDir()
	dataDir := filepath.Join(workDir, "runtime", "data")
	repoDataDir := filepath.Join(workDir, "repo", "store", "data")
	require.NoError(t, os.MkdirAll(dataDir, 0o755))
	m := NewFileSystemManager(dataDir, "", "", "")

	transcriptionPath := filepath.Join(repoDataDir, "transcriptions", "Paris_1615", "page-0001", "original.txt")
	require.NoError(t, os.MkdirAll(filepath.Dir(transcriptionPath), 0o755))
	require.NoError(t, os.WriteFile(transcriptionPath, []byte("text"), 0o644))
	require.NoError(t, os.Symlink(filepath.Join(repoDataDir, "transcriptions"), filepath.Join(dataDir, "transcriptions")))

	toDelete, err := m.CleanupLocalStore(false, nil, nil)
	require.NoError(t, err)
	require.NotContains(t, toDelete, filepath.Join(dataDir, "transcriptions"))
	require.FileExists(t, transcriptionPath)
	require.FileExists(t, filepath.Join(dataDir, "transcriptions", "Paris_1615", "page-0001", "original.txt"))
}

func TestCleanupLocalStoreDeletesDatasetFilesByAbsolutePath(t *testing.T) {
	workDir := t.TempDir()
	dataDir := filepath.Join(workDir, "data")
	require.NoError(t, os.MkdirAll(dataDir, 0o755))
	t.Chdir(workDir)

	m := NewFileSystemManager(dataDir, "", "", "")
	datasetDir := filepath.Join(dataDir, "ds_keep")
	strayDatasetFile := filepath.Join(datasetDir, "unexpected.txt")
	cwdSameNamedFile := filepath.Join(workDir, "unexpected.txt")

	require.NoError(t, os.MkdirAll(filepath.Join(datasetDir, "imgs"), 0o755))
	require.NoError(t, os.WriteFile(strayDatasetFile, []byte("delete me"), 0o644))
	require.NoError(t, os.WriteFile(cwdSameNamedFile, []byte("keep me"), 0o644))

	toDelete, err := m.CleanupLocalStore(false, map[string][]*annotation.Annotation{}, []*model.Dataset{
		{Meta: common.NewMeta("ds_keep")},
	})
	require.NoError(t, err)
	require.Contains(t, toDelete, strayDatasetFile)
	require.NoFileExists(t, strayDatasetFile)
	require.FileExists(t, cwdSameNamedFile)
}

func TestCleanupLocalStorePreservesAllowedDatasetSymlinks(t *testing.T) {
	workDir := t.TempDir()
	dataDir := filepath.Join(workDir, "runtime", "data")
	repoDataDir := filepath.Join(workDir, "repo", "store", "data")
	require.NoError(t, os.MkdirAll(dataDir, 0o755))
	m := NewFileSystemManager(dataDir, "", "", "")

	datasetDir := filepath.Join(dataDir, "tps")
	imgsTarget := filepath.Join(repoDataDir, "tps", "imgs")
	imgPath := filepath.Join(imgsTarget, "Paris_1622_tp.jpeg")
	require.NoError(t, os.MkdirAll(imgsTarget, 0o755))
	require.NoError(t, os.WriteFile(imgPath, []byte("image"), 0o644))
	require.NoError(t, os.MkdirAll(datasetDir, 0o755))
	require.NoError(t, os.Symlink(imgsTarget, filepath.Join(datasetDir, "imgs")))

	toDelete, err := m.CleanupLocalStore(false, map[string][]*annotation.Annotation{}, []*model.Dataset{
		{Meta: common.NewMeta("tps")},
	})
	require.NoError(t, err)
	require.NotContains(t, toDelete, filepath.Join(datasetDir, "imgs"))
	require.FileExists(t, filepath.Join(datasetDir, "imgs", "Paris_1622_tp.jpeg"))
}
