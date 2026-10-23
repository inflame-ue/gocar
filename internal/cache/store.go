package cache

import (
	"path/filepath"
	"strings"
)

func safeStore(root, path string) (string, error) {
	name := filepath.ToSlash(path)
	if len(path) == 0 {
		return "", &PathError{
			Path: path,
			Msg:  "path cannot be empty",
		}
	}

	name = filepath.Clean(name)
	if filepath.IsAbs(name) {
		return "", &PathError{
			Path: path,
			Msg:  "path cannot be absolute",
		}
	}

	rel, err := filepath.Rel(root, filepath.Join(root, name))
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", &PathError{
			Path: path,
			Msg:  "path escapes the root",
		}
	}

	return name, nil
}
