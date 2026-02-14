package unit_test_scaffolder

import (
	"os"
	"path/filepath"
	"strings"
)

// Scanner identifies relevant files for unit testing.
type Scanner struct {
	SourceDir string
}

// Scan finds candidate files for test generation.
func (s *Scanner) Scan(specified []string) ([]string, error) {
	if len(specified) > 0 {
		return specified, nil
	}

	var candidates []string
	err := filepath.Walk(s.SourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Look for common business logic files (excluding tests)
		name := strings.ToLower(info.Name())
		if strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, ".test.ts") || strings.Contains(path, "vendor") {
			return nil
		}

		// Filter by common patterns
		isCandidate := strings.Contains(name, "controller") ||
			strings.Contains(name, "service") ||
			strings.Contains(name, "repo") ||
			strings.Contains(name, "handler") ||
			strings.Contains(name, "usecase")

		if isCandidate {
			candidates = append(candidates, path)
		}

		return nil
	})

	return candidates, err
}
