package cache

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func safeStore(root, path string) (string, error) {
	name := filepath.ToSlash(path)
	if len(name) == 0 {
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

func safeJoin(root, path string) (string, error) {
	file, err := safeStore(root, path)
	if err != nil {
		return "", err
	}

	return filepath.Join(root, file), nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return &IOError{
			Path:  src,
			Cause: err,
		}
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return &IOError{
			Path:  dst,
			Cause: err,
		}
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return &IOError{Path: dst, Cause: err}
	}
	return nil
}
