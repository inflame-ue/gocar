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

	// do not allow absolute paths
	if strings.HasPrefix(clean, "/") || filepath.IsAbs(clean) {
		return nil, &PatternError{
			Pattern: clean,
			Msg:     "absolute paths are not allowed",
		}
	}

	segments := strings.Split(clean, "/")
	for _, segment := range segments {
		// do not allow ** inside a segment
		if strings.Contains(segment, "**") && segment != "**" {
			return nil, &PatternError{
				Pattern: clean,
				Msg:     "** inside segments are not allowed",
			}
		}

		// do not allow escaping patterns like ../foo
		if segment == ".." {
			return nil, &PatternError{
				Pattern: clean,
				Msg:     "escaping patterns are not allowed",
			}
		}
	}

	return segments, nil
}
