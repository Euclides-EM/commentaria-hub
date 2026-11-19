package diagramcrops

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/config"
)

type volumeInfo struct {
	Key    string
	Volume int
}

type volumeData struct {
	Volume      int      `json:"volume,omitempty"`
	Key         string   `json:"key,omitempty"`
	Images      []string `json:"images"`
	HasDiagrams bool     `json:"hasDiagrams"`
}

type multiVolumeData struct {
	Volumes []volumeData `json:"volumes"`
}

// Options configures diagram crops generation.
type Options struct {
	DryRun bool
	Force  bool
}

type generator struct {
	diagramsLocalDir  string
	itemsMetadataDir  string
	diagramsSourceDir string
	opts              Options
}

// Generate reads diagram crop image names from FACSIMILES_DIAGRAMS_PATH and writes small metadata JSON
// files under the app store. The source path must be an absolute local path or file:// URL.
func Generate(env *config.EnvConfig, opts Options) error {
	if env.SkipDiagramCropsGeneration {
		log.Printf("diagram crops: skipping diagram metadata generation because SKIP_DIAGRAM_CROPS_GENERATION is set")
		return nil
	}
	return GenerateFromPaths(env.FacsimilesDiagramsPath, env.DiagramsDir(), env.ItemsMetadataStoreDir(), opts)
}

func GenerateFromPaths(diagramsPath, diagramsLocalDir, itemsMetadataDir string, opts Options) error {
	diagramsSourceDir, err := localDiagramsSourceDir(diagramsPath)
	if err != nil {
		return err
	}
	if diagramsSourceDir == "" {
		log.Printf("diagram crops: FACSIMILES_DIAGRAMS_PATH is not an absolute local path, skipping diagram metadata generation")
		return nil
	}

	g := &generator{
		diagramsLocalDir:  diagramsLocalDir,
		itemsMetadataDir:  itemsMetadataDir,
		diagramsSourceDir: diagramsSourceDir,
		opts:              opts,
	}

	directories, err := g.generateDiagramDirectories()
	if err != nil {
		return err
	}

	return g.generateAllDiagramData(directories)
}

func localDiagramsSourceDir(pathOrURL string) (string, error) {
	pathOrURL = strings.TrimSpace(pathOrURL)
	if pathOrURL == "" {
		return "", nil
	}
	parsed, err := url.Parse(pathOrURL)
	if err == nil && parsed.Scheme == "file" {
		if parsed.Host != "" && !strings.EqualFold(parsed.Host, "localhost") {
			return "", fmt.Errorf("unsupported file URL host for diagrams path: %s", parsed.Host)
		}
		return parsed.Path, nil
	}
	if filepath.IsAbs(pathOrURL) {
		return pathOrURL, nil
	}
	return "", nil
}

func (g *generator) writeJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')

	if g.opts.DryRun {
		log.Printf("dry-run: would write %s (%d bytes)", path, len(b))
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, b, 0o644)
}

func (g *generator) generateDiagramDirectories() ([]string, error) {
	log.Printf("fetching diagram directories from local path %s", g.diagramsSourceDir)
	entries, err := os.ReadDir(g.diagramsSourceDir)
	if err != nil {
		return nil, err
	}
	directories := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}
	slices.Sort(directories)

	outputPath := filepath.Join(g.itemsMetadataDir, "diagram-directories.json")
	if err := g.writeJSON(outputPath, directories); err != nil {
		return nil, err
	}

	log.Printf("generated %s with %d directories", outputPath, len(directories))
	return directories, nil
}

func groupDirectoriesByBase(directories []string) map[string][]volumeInfo {
	grouped := make(map[string][]volumeInfo, len(directories))

	for _, dir := range directories {
		if matches := volSuffixRe.FindStringSubmatch(dir); len(matches) == 3 {
			baseKey := matches[1]
			volNum, _ := strconv.Atoi(matches[2])
			grouped[baseKey] = append(grouped[baseKey], volumeInfo{Key: dir, Volume: volNum})
			continue
		}
		grouped[dir] = append(grouped[dir], volumeInfo{Key: dir, Volume: 1})
	}

	for baseKey := range grouped {
		slices.SortFunc(grouped[baseKey], func(a, b volumeInfo) int { return a.Volume - b.Volume })
	}

	return grouped
}

func (g *generator) generateAllDiagramData(directories []string) error {
	grouped := groupDirectoriesByBase(directories)

	missing := make([]string, 0, len(grouped))
	for baseKey := range grouped {
		outputPath := filepath.Join(g.diagramsLocalDir, baseKey+".json")
		if _, err := os.Stat(outputPath); os.IsNotExist(err) || g.opts.DryRun || g.opts.Force {
			missing = append(missing, baseKey)
		}
	}

	if len(missing) == 0 {
		log.Printf("all diagram data files already exist, skipping")
		return nil
	}

	slices.Sort(missing)
	log.Printf("generating diagram data for %d of %d base entries", len(missing), len(grouped))

	for i, baseKey := range missing {
		if err := g.generateDiagramData(baseKey, grouped[baseKey], fmt.Sprintf("[%d/%d]", i, len(missing))); err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

func (g *generator) generateDiagramData(baseKey string, volumes []volumeInfo, logPrefix string) error {
	outputPath := filepath.Join(g.diagramsLocalDir, baseKey+".json")
	if _, err := os.Stat(outputPath); err == nil && !g.opts.DryRun && !g.opts.Force {
		return nil
	}

	volumeRows := make([]volumeData, 0, len(volumes))
	for _, v := range volumes {
		images, err := g.fetchCrops(v.Key)
		row := volumeData{
			Key:         v.Key,
			Images:      images,
			HasDiagrams: len(images) != 0,
		}
		if len(volumes) > 1 {
			row.Volume = v.Volume
		}
		if err != nil {
			row.Images = []string{}
			row.HasDiagrams = false
		}
		volumeRows = append(volumeRows, row)
		time.Sleep(100 * time.Millisecond)
	}

	var out any
	if len(volumes) > 1 {
		out = multiVolumeData{Volumes: volumeRows}
	} else if len(volumeRows) > 0 {
		out = volumeRows[0]
	} else {
		out = volumeData{Images: []string{}, HasDiagrams: false}
	}

	if err := g.writeJSON(outputPath, out); err != nil {
		return err
	}

	totalImages := 0
	for _, row := range volumeRows {
		totalImages += len(row.Images)
	}
	log.Printf("%s generated diagrams data for %s: %d images across %d volume(s)", logPrefix, baseKey, totalImages, len(volumes))

	return nil
}

func (g *generator) fetchCrops(volumeKey string) ([]string, error) {
	cropsDir := filepath.Join(g.diagramsSourceDir, volumeKey, "crops")
	items, err := os.ReadDir(cropsDir)
	if err != nil {
		return nil, err
	}
	images := make([]string, 0, len(items))
	for _, item := range items {
		if !item.IsDir() && strings.HasSuffix(item.Name(), ".jpg") {
			images = append(images, item.Name())
		}
	}

	slices.Sort(images)
	return images, nil
}
