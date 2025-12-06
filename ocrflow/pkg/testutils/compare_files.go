package testutils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CompareDirs checks that two directory trees are identical
// in structure and in file contents.
func CompareDirs(dirA, dirB string, contentPhrasesSwitch map[string]string) error {
	return filepath.Walk(dirA, func(pathA string, infoA os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk error: %w", err)
		}

		// Compute the equivalent path in dirB
		rel, err := filepath.Rel(dirA, pathA)
		if err != nil {
			return fmt.Errorf("rel error: %w", err)
		}
		pathB := filepath.Join(dirB, rel)

		infoB, err := os.Stat(pathB)
		if err != nil {
			return fmt.Errorf("missing in %s: %s", dirB, pathB)
		}

		// Type mismatch
		if infoA.IsDir() != infoB.IsDir() {
			return fmt.Errorf("type mismatch for %s", rel)
		}

		// If directory, nothing else to check
		if infoA.IsDir() {
			return nil
		}

		// Compare contents
		same, err := filesEqual(pathA, pathB, contentPhrasesSwitch)
		if err != nil {
			return err
		}
		if !same {
			return fmt.Errorf("file mismatch: %s vs %s", pathA, pathB)
		}

		return nil
	})
}

// filesEqual compares two files byte-for-byte
func filesEqual(a, b string, phrasesSwitch map[string]string) (bool, error) {
	f1, err := os.Open(a)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(b)
	if err != nil {
		return false, err
	}
	defer f2.Close()

	f1Content, err := io.ReadAll(f1)
	if err != nil {
		return false, err
	}
	f2Content, err := io.ReadAll(f2)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(string(f1Content)) == strings.TrimSpace(string(f2Content)) {
		return true, nil
	}

	for k, v := range phrasesSwitch {
		f1Content = bytes.ReplaceAll(f1Content, []byte(k), []byte(v))
	}

	if strings.TrimSpace(string(f1Content)) == strings.TrimSpace(string(f2Content)) {
		return true, nil
	}

	return false, nil
}
