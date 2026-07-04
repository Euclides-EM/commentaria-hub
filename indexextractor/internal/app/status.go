package app

import (
	"fmt"
	"io"
)

func reportStatus(cfg config, out io.Writer) error {
	if includesKind(cfg.kind, kindIndex) {
		images, err := discoverImages(cfg.indexDir)
		if err != nil {
			return fmt.Errorf("status index: %w", err)
		}
		manifest, err := loadIndexManifest(cfg.indexCSV, true)
		if err != nil {
			return err
		}
		reportCompletion("index", images, indexCompleted(manifest), cfg.indexCSV, out)
		reportIssues("index", indexManifestIssues(manifest), out)
		reportFailures("index", indexFailures(manifest), out)
	}
	if includesKind(cfg.kind, kindLetters) {
		images, err := discoverImages(cfg.lettersDir)
		if err != nil {
			return fmt.Errorf("status letters: %w", err)
		}
		manifest, err := loadLettersManifest(cfg.lettersCSV, true)
		if err != nil {
			return err
		}
		reportCompletion("letters", images, lettersCompleted(manifest), cfg.lettersCSV, out)
		reportIssues("letters", lettersManifestIssues(manifest), out)
		reportFailures("letters", lettersFailures(manifest), out)
	}
	return nil
}

func reportCompletion(label string, images []imageInput, completed map[string]bool, outputPath string, out io.Writer) {
	known := 0
	for _, image := range images {
		if completed[cleanPath(image.Path)] {
			known++
		}
	}
	fmt.Fprintf(out, "%s: %d/%d images completed, %d pending; manifest %s\n", label, known, len(images), len(images)-known, manifestPath(outputPath))
}

func indexManifestIssues(manifest indexManifest) []pageIssues {
	pages := make([]pageIssues, 0)
	for _, page := range manifest.Pages {
		if len(page.Issues) > 0 {
			pages = append(pages, pageIssues{ImagePath: page.ImagePath, Issues: page.Issues})
		}
	}
	return pages
}

func lettersManifestIssues(manifest lettersManifest) []pageIssues {
	pages := make([]pageIssues, 0)
	for _, page := range manifest.Pages {
		if len(page.Issues) > 0 {
			pages = append(pages, pageIssues{ImagePath: page.ImagePath, Issues: page.Issues})
		}
	}
	return pages
}

func reportIssues(label string, pages []pageIssues, out io.Writer) {
	total := 0
	for _, page := range pages {
		total += len(page.Issues)
	}
	fmt.Fprintf(out, "\n%s: %d tolerated parsing issues across %d affected images\n", label, total, len(pages))
	for _, page := range pages {
		for _, issue := range page.Issues {
			fmt.Fprintf(out, "  %s: %s\n", page.ImagePath, issue)
		}
	}
}

func indexFailures(manifest indexManifest) []failedPage {
	pages := make([]failedPage, 0)
	for _, page := range manifest.Pages {
		if page.Failure != nil {
			pages = append(pages, failedPage{ImagePath: page.ImagePath, Failure: page.Failure})
		}
	}
	return pages
}

func lettersFailures(manifest lettersManifest) []failedPage {
	pages := make([]failedPage, 0)
	for _, page := range manifest.Pages {
		if page.Failure != nil {
			pages = append(pages, failedPage{ImagePath: page.ImagePath, Failure: page.Failure})
		}
	}
	return pages
}

func reportFailures(label string, pages []failedPage, out io.Writer) {
	fmt.Fprintf(out, "\n%s: %d failed images\n", label, len(pages))
	for _, page := range pages {
		fmt.Fprintf(out, "  %s: %s via %s/%s: %s\n", page.ImagePath, page.Failure.Phase, page.Failure.Provider, page.Failure.Model, page.Failure.Error)
	}
}
