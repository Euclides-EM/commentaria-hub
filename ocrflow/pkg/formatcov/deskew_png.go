package formatcov

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
	"golang.org/x/sync/errgroup"
)

// DeskewOptions are optional parameters for the deskew CLI (one input file per call; no folder support).
// See: deskew --help (--sigma, --num-peaks, --num-angles, --background).
type DeskewOptions struct {
	Sigma      *float64 // --sigma
	NumPeaks   *int     // --num-peaks
	NumAngles  *int     // --num-angles
	Background *string  // --background
}

// maxDeskewWorkers limits concurrent deskew processes (CPU-bound + subprocess).
func maxDeskewWorkers() int {
	n := runtime.NumCPU()
	if n < 2 {
		return 2
	}
	if n > 24 {
		return 24
	}
	return n
}

func DeskewPNGs(src string, dst string) error {
	return DeskewPNGsWithOptions(src, dst, nil)
}

// DeskewPNGsWithOptions deskews all PNGs in src into dst with optional CLI flags.
// The deskew CLI accepts only one input file per invocation (no folder/multi-file), so we iterate and run in parallel.
func DeskewPNGsWithOptions(src string, dst string, opts *DeskewOptions) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read src dir %q: %w", src, err)
	}

	var jobs []struct{ name, inPath, outPath string }
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.ToLower(filepath.Ext(name)) != ".png" {
			continue
		}
		jobs = append(jobs, struct{ name, inPath, outPath string }{
			name:    name,
			inPath:  filepath.Join(src, name),
			outPath: filepath.Join(dst, name),
		})
	}

	if len(jobs) == 0 {
		return nil
	}

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create dst dir %q: %w", dst, err)
	}

	workers := maxDeskewWorkers()
	log.Printf("Deskewing %d images with %d workers", len(jobs), workers)

	sem := make(chan struct{}, workers)
	grp, ctx := errgroup.WithContext(context.Background())

	for i, job := range jobs {
		i, job := i, job
		grp.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sem <- struct{}{}:
				defer func() { <-sem }()
			}
			log.Printf("[%d/%d] Deskewing %q -> %q", i+1, len(jobs), job.inPath, job.outPath)
			args := deskewArgs(job.outPath, job.inPath, opts)
			if err := envexec.PythonCmd("deskew", args...); err != nil {
				return fmt.Errorf("deskew image %q: %w", job.inPath, err)
			}
			return nil
		})
	}

	if err := grp.Wait(); err != nil {
		return err
	}
	return nil
}

func deskewArgs(outPath, inPath string, opts *DeskewOptions) []string {
	args := []string{"--output", outPath}
	if opts != nil {
		if opts.Sigma != nil {
			args = append(args, "--sigma", strconv.FormatFloat(*opts.Sigma, 'f', -1, 64))
		}
		if opts.NumPeaks != nil {
			args = append(args, "--num-peaks", strconv.Itoa(*opts.NumPeaks))
		}
		if opts.NumAngles != nil {
			args = append(args, "--num-angles", strconv.Itoa(*opts.NumAngles))
		}
		if opts.Background != nil {
			args = append(args, "--background", *opts.Background)
		}
	}
	args = append(args, inPath)
	return args
}
