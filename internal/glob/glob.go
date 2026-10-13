package glob

import (
	"path/filepath"
	"strings"
)

func Expand(dir, pattern string) ([]string, error) {
	return nil, nil
}

func splitSegments(pattern string) ([]string, error) {
	clean := filepath.Clean(filepath.ToSlash(pattern))

	if strings.HasPrefix(clean, "/") || filepath.IsAbs(clean) {
		return nil, &NoMatchError{
			Pattern: clean,
		}
	}

	segments := filepath.SplitList(clean)
	for _, segment := range segments {
		if strings.Contains(segment, "**") && segment != "**" {
			return nil, &NoMatchError{
				Pattern: clean,
			}
		}
	}

	return segments, nil
}
