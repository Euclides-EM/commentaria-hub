package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	githubAPIBase = "https://api.github.com/repos/Euclides-EM/elements-facsimile/contents"
	diagramsPath  = "docs/diagrams"
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
	Volume        int      `json:"volume,omitempty"`
	Key           string   `json:"key,omitempty"`
	Images        []string `json:"images"`
	HasNoDiagrams bool     `json:"hasNoDiagrams"`
}

type multiVolumeData struct {
	Volumes []volumeData `json:"volumes"`
}

type generator struct {
	client    *http.Client
	headers   map[string]string
	outputDir string
	dryRun    bool
}

func main() {
	var dryRun bool
	flag.BoolVar(&dryRun, "dry-run", false, "print actions without writing files")
	flag.Parse()

	repoRoot, err := findRepoRoot()
	if err != nil {
		log.Fatalf("failed to locate repo root: %v", err)
	}

	githubPAT, err := loadGitHubPAT(repoRoot)
	if err != nil {
		log.Fatalf("failed to load github pat: %v", err)
	}

	g := &generator{
		client:    &http.Client{Timeout: 20 * time.Second},
		headers:   buildHeaders(githubPAT),
		outputDir: filepath.Join(repoRoot, "store", "items_metadata"),
		dryRun:    dryRun,
	}

	directories, err := g.generateDiagramDirectories()
	if err != nil {
		log.Fatalf("failed to generate diagram directories: %v", err)
	}

	if err := g.generateAllDiagramData(directories); err != nil {
		log.Fatalf("failed to generate diagram data: %v", err)
	}
}

func buildHeaders(githubPAT string) map[string]string {
	headers := map[string]string{
		"User-Agent":    "elements-title-pages-build",
		"Authorization": "token " + githubPAT,
	}
	return headers
}

func loadGitHubPAT(repoRoot string) (string, error) {
	parentRoot := filepath.Dir(repoRoot)
	envFiles := uniquePaths([]string{
		filepath.Join(repoRoot, ".env"),
		filepath.Join(repoRoot, ".env_private"),
		filepath.Join(parentRoot, ".env"),
		filepath.Join(parentRoot, ".env_private"),
	})

	var githubToken string
	for _, envFile := range envFiles {
		envMap, err := godotenv.Read(envFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", fmt.Errorf("failed to read %s: %w", envFile, err)
		}
		if pat := strings.TrimSpace(envMap["GITHUB_PAT"]); pat != "" {
			githubToken = pat
			continue
		}
		if token := strings.TrimSpace(envMap["GITHUB_TOKEN"]); token != "" {
			githubToken = token
		}
	}

	if githubToken == "" {
		return "", fmt.Errorf("GITHUB_PAT or GITHUB_TOKEN is required in one of: %s", strings.Join(envFiles, ", "))
	}

	return githubToken, nil
}

func uniquePaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := wd
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}

		next := filepath.Dir(current)
		if next == current {
			return "", errors.New("go.mod not found")
		}
		current = next
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
		lastErr = errors.New("request failed")
	}
	return nil, lastErr
}

func (g *generator) fetchGithubContents(path string) ([]githubContent, error) {
	url := fmt.Sprintf("%s/%s", githubAPIBase, path)
	body, err := g.fetchWithRetry(url, 3)
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

	if g.dryRun {
		log.Printf("dry-run: would write %s (%d bytes)", path, len(b))
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, b, 0o644)
}

func (g *generator) generateDiagramDirectories() ([]string, error) {
	log.Printf("fetching diagram directories from GitHub API")
	items, err := g.fetchGithubContents(diagramsPath)
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

	outputPath := filepath.Join(g.outputDir, "diagram-directories.json")
	if err := g.writeJSON(outputPath, directories); err != nil {
		return nil, err
	}

	log.Printf("generated %s with %d directories", outputPath, len(directories))
	return directories, nil
}

func groupDirectoriesByBase(directories []string) map[string][]volumeInfo {
	grouped := make(map[string][]volumeInfo, len(directories))
	re := regexp.MustCompile(`^(.+)_vol(\\d+)$`)

	for _, dir := range directories {
		if matches := re.FindStringSubmatch(dir); len(matches) == 3 {
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
		outputPath := filepath.Join(g.outputDir, "diagrams", baseKey+".json")
		if _, err := os.Stat(outputPath); errors.Is(err, os.ErrNotExist) || g.dryRun {
			missing = append(missing, baseKey)
		}
	}

	if len(missing) == 0 {
		log.Printf("all diagram data files already exist, skipping")
		return nil
	}

	slices.Sort(missing)
	log.Printf("generating diagram data for %d of %d base entries", len(missing), len(grouped))

	for _, baseKey := range missing {
		if err := g.generateDiagramData(baseKey, grouped[baseKey]); err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

func (g *generator) generateDiagramData(baseKey string, volumes []volumeInfo) error {
	outputPath := filepath.Join(g.outputDir, "diagrams", baseKey+".json")
	if _, err := os.Stat(outputPath); err == nil && !g.dryRun {
		return nil
	}

	volumeRows := make([]volumeData, 0, len(volumes))
	for _, v := range volumes {
		images, err := g.fetchCrops(v.Key)
		row := volumeData{
			Key:           v.Key,
			Images:        images,
			HasNoDiagrams: len(images) == 0,
		}
		if len(volumes) > 1 {
			row.Volume = v.Volume
		}
		if err != nil {
			row.Images = []string{}
			row.HasNoDiagrams = true
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
		out = volumeData{Images: []string{}, HasNoDiagrams: true}
	}

	if err := g.writeJSON(outputPath, out); err != nil {
		return err
	}

	totalImages := 0
	for _, row := range volumeRows {
		totalImages += len(row.Images)
	}
	log.Printf("generated diagrams data for %s: %d images across %d volume(s)", baseKey, totalImages, len(volumes))

	return nil
}

func (g *generator) fetchCrops(volumeKey string) ([]string, error) {
	items, err := g.fetchGithubContents(filepath.Join(diagramsPath, volumeKey, "crops"))
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
