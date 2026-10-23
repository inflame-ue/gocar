package glob

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

func Expand(dir, pattern string) ([]string, error) {
	segments, err := splitSegments(pattern)
	if err != nil {
		return nil, err
	}

	hits := []string{}
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			return nil
		}

		ok, err := match(segments, strings.Split(rel, "/"))
		if err != nil {
			return &PatternError{
				Pattern: pattern, 
				Msg: err.Error(),
			}
		}

		if ok {
			hits = append(hits, rel)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(hits) == 0 {
		return nil, &NoMatchError{
			Pattern: pattern,
		}
	}

	slices.Sort(hits)
	return hits, nil
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
