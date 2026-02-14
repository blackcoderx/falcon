package shared

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidatePathWithinWorkDir checks if a given path is within the allowed work directory.
// This prevents path traversal attacks (e.g., "../../../etc/passwd").
func ValidatePathWithinWorkDir(filePath, workDir string) (absPath string, err error) {
	targetPath := filePath
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(workDir, targetPath)
	}

	absPath, err = filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve work directory: %w", err)
	}

	if !strings.HasSuffix(absWorkDir, string(filepath.Separator)) {
		absWorkDir += string(filepath.Separator)
	}

	if absPath != strings.TrimSuffix(absWorkDir, string(filepath.Separator)) &&
		!strings.HasPrefix(absPath, absWorkDir) {
		return "", fmt.Errorf("access denied: path outside project directory")
	}

	return absPath, nil
}
