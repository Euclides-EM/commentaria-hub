package diagramcrops

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/config"
)

type githubContent struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

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
}

type generator struct {
	client             *http.Client
	headers            map[string]string
	diagramsLocalDir   string
	itemsMetadataDir   string
	diagramsRemotePath string
	diagramsSourceDir  string
	githubAPIBase      string
	opts               Options
}

// ResolveGithubAPIBase converts a repo URL (e.g. https://github.com/org/repo) into the GitHub API contents base URL.
func ResolveGithubAPIBase(repoURL string) (string, error) {
	repoURL = strings.TrimSuffix(strings.TrimSpace(repoURL), "/")
	parsed, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("parse facsimiles repo url: %w", err)
	}

	if strings.EqualFold(parsed.Host, "api.github.com") {
		return repoURL, nil
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if strings.EqualFold(parsed.Host, "github.com") {
		if len(parts) < 2 {
			return "", fmt.Errorf("unsupported github URL format: %s", repoURL)
		}
		return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", parts[0], parts[1]), nil
	}

	if strings.EqualFold(parsed.Host, "raw.githubusercontent.com") {
		if len(parts) < 3 {
			return "", fmt.Errorf("unsupported raw github URL format: %s", repoURL)
		}
		return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", parts[0], parts[1]), nil
	}

	return "", fmt.Errorf("unsupported host for facsimiles repo url: %s", repoURL)
}

// Generate fetches diagram directory listing and per-edition diagram data from the facsimiles GitHub repo
// and writes them under env's store dir. It uses env for diagram path, GitHub token, and store paths.
// If EffectiveGithubToken is empty, Generate returns nil without doing anything (allows app to start without token).
func Generate(env *config.EnvConfig, opts Options) error {
	if env.SkipDiagramCropsGeneration {
		log.Printf("diagram crops: skipping diagram metadata generation because SKIP_DIAGRAM_CROPS_GENERATION is set")
		return nil
	}
	diagramsSourceDir, err := localDiagramsSourceDir(env.FacsimilesDiagramsPath)
	if err != nil {
		return err
	}
	token := env.GithubToken
	if diagramsSourceDir == "" && token == "" {
		log.Printf("diagram crops: no GITHUB_TOKEN set, skipping diagram metadata generation")
		return nil
	}

	githubAPIBase := ""
	if diagramsSourceDir == "" {
		githubAPIBase, err = ResolveGithubAPIBase(env.FacsimilesGithubRepoUrl)
		if err != nil {
			return fmt.Errorf("resolve github api base: %w", err)
		}
	}

	g := &generator{
		client:             &http.Client{Timeout: 20 * time.Second},
		headers:            buildHeaders(token),
		diagramsLocalDir:   env.DiagramsDir(),
		itemsMetadataDir:   env.ItemsMetadataStoreDir(),
		diagramsRemotePath: strings.TrimSuffix(env.FacsimilesDiagramsPath, "/"),
		diagramsSourceDir:  diagramsSourceDir,
		githubAPIBase:      githubAPIBase,
		opts:               opts,
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

func buildHeaders(githubToken string) map[string]string {
	return map[string]string{
		"User-Agent":    "elements-title-pages-build",
		"Authorization": "token " + githubToken,
	}
}

func (g *generator) fetchWithRetry(url string, retries int) ([]byte, error) {
	var lastErr error

	for i := 0; i < retries; i++ {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		for k, v := range g.headers {
			req.Header.Set(k, v)
		}

		resp, err := g.client.Do(req)
		if err != nil {
			lastErr = err
			if i < retries-1 {
				time.Sleep(time.Second)
				continue
			}
			break
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if i < retries-1 {
				time.Sleep(time.Second)
				continue
			}
			break
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return body, nil
		}

		if resp.StatusCode == http.StatusForbidden {
			remaining := resp.Header.Get("X-RateLimit-Remaining")
			resetRaw := resp.Header.Get("X-RateLimit-Reset")
			log.Printf("rate limit hit (%s remaining), retrying in %d seconds", remaining, 1<<i)
			if resetRaw != "" {
				if unix, err := strconv.ParseInt(resetRaw, 10, 64); err == nil {
					log.Printf("rate limit resets at: %s", time.Unix(unix, 0).UTC().Format(time.RFC3339))
				}
			}
			if i < retries-1 {
				time.Sleep(time.Duration(1<<i) * time.Second)
				continue
			}
		}

		lastErr = fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		if i < retries-1 {
			time.Sleep(time.Second)
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("request failed")
	}
	return nil, lastErr
}

func (g *generator) fetchGithubContents(path string) ([]githubContent, error) {
	contentURL := fmt.Sprintf("%s/%s", g.githubAPIBase, path)
	body, err := g.fetchWithRetry(contentURL, 3)
	if err != nil {
		return nil, err
	}

	var items []githubContent
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}

	return items, nil
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
	if g.diagramsSourceDir != "" {
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

	log.Printf("fetching diagram directories from GitHub API")
	items, err := g.fetchGithubContents(g.diagramsRemotePath)
	if err != nil {
		return nil, err
	}

	directories := make([]string, 0, len(items))
	for _, item := range items {
		if item.Type == "dir" {
			directories = append(directories, item.Name)
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

var volSuffixRe = regexp.MustCompile(`^(.+)_vol([0-9]+)$`)

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
		if _, err := os.Stat(outputPath); os.IsNotExist(err) || g.opts.DryRun {
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
	if _, err := os.Stat(outputPath); err == nil && !g.opts.DryRun {
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
	if g.diagramsSourceDir != "" {
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

	path := g.diagramsRemotePath + "/" + volumeKey + "/crops"
	items, err := g.fetchGithubContents(path)
	if err != nil {
		return nil, err
	}

	images := make([]string, 0, len(items))
	for _, item := range items {
		if strings.HasSuffix(item.Name, ".jpg") {
			images = append(images, item.Name)
		}
	}

	slices.Sort(images)
	return images, nil
}
