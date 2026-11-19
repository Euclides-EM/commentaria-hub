package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
)

func main() {
	var reassignmentFlags repeatedFlag
	segmentationDir := flag.String("segmentation-dir", "", "flat directory containing zone/segmentation ALTO files")
	ocrDir := flag.String("ocr-dir", "", "directory containing page-NNNN/original.xml OCR ALTO files (flat XML is also accepted)")
	outputDir := flag.String("output-dir", "", "directory for merged page-NNNN/original.xml files")
	include := flag.String("include-categories", "", "comma-separated category labels allowed to contain text")
	ignore := flag.String("ignore-categories", "", "comma-separated category labels that cut holes out of included zones")
	flag.Var(&reassignmentFlags, "reassign", "ordered line reassignment as FROM:TO:PRECISION_PX:MIN_OVERLAP; may be repeated")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s -segmentation-dir DIR -ocr-dir DIR -output-dir DIR -include-categories LABELS [-ignore-categories LABELS]\n\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	stats, err := run(*segmentationDir, *ocrDir, *outputDir, *include, *ignore, reassignmentFlags)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("merged %d pages: %d input lines -> %d output fragments; removed %d lines and %d OCR characters; split/clipped %d lines; reassigned %d lines\n",
		stats.Pages, stats.InputLines, stats.OutputLines, stats.RemovedLines, stats.RemovedRunes, stats.SplitLines, stats.ReassignedLines)
}

func run(segmentationDir, ocrDir, outputDir, include, ignore string, reassignmentFlags []string) (alto.MergeStats, error) {
	if strings.TrimSpace(segmentationDir) == "" {
		return alto.MergeStats{}, errors.New("missing -segmentation-dir")
	}
	if strings.TrimSpace(ocrDir) == "" {
		return alto.MergeStats{}, errors.New("missing -ocr-dir")
	}
	if strings.TrimSpace(outputDir) == "" {
		return alto.MergeStats{}, errors.New("missing -output-dir")
	}
	reassignments := make([]alto.LineReassignment, 0, len(reassignmentFlags))
	for i, value := range reassignmentFlags {
		rule, err := parseReassignment(value)
		if err != nil {
			return alto.MergeStats{}, fmt.Errorf("invalid -reassign value %d: %w", i+1, err)
		}
		reassignments = append(reassignments, rule)
	}
	return alto.MergeZonedOCRDirsWithReassignments(segmentationDir, ocrDir, outputDir, splitCategories(include), splitCategories(ignore), reassignments)
}

type repeatedFlag []string

func (f *repeatedFlag) String() string { return strings.Join(*f, ", ") }
func (f *repeatedFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func parseReassignment(value string) (alto.LineReassignment, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 4 {
		return alto.LineReassignment{}, fmt.Errorf("expected FROM:TO:PRECISION_PX:MIN_OVERLAP, got %q", value)
	}
	precision, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err != nil {
		return alto.LineReassignment{}, fmt.Errorf("invalid precision_px %q: %w", parts[2], err)
	}
	minOverlap, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
	if err != nil {
		return alto.LineReassignment{}, fmt.Errorf("invalid min_overlap %q: %w", parts[3], err)
	}
	return alto.LineReassignment{
		FromCategory: strings.TrimSpace(parts[0]),
		ToCategory:   strings.TrimSpace(parts[1]),
		PrecisionPx:  precision,
		MinOverlap:   minOverlap,
	}, nil
}

func splitCategories(value string) []string {
	var result []string
	for _, category := range strings.Split(value, ",") {
		if category = strings.TrimSpace(category); category != "" {
			result = append(result, category)
		}
	}
	return result
}
