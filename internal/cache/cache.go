package cache

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

type L1 struct {
	root string
}

func NewL1(root string) *L1 {
	return &L1{
		root: root,
	}
}

func (c *L1) Put(key string, outputs []string, dir string) error {
	err := os.MkdirAll(c.root, 0o700)
	if err != nil {
		return &IOError{Path: c.root, Cause: err}
	}

	cachePath := filepath.Join(c.root, key)

	// the cachePath is already present at c.root --> return
	if _, err := os.Stat(cachePath); err == nil {
		return nil
	}

	tmp, err := os.MkdirTemp(c.root, "tmp-*")
	if err != nil {
		return &IOError{Path: cachePath, Cause: err}
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	// note: if outputs slice is empty, the loop is skipped
	// we rename the tmp directory without copying any bytes
	for _, path := range outputs {
		dest, err := safeJoin(tmp, path)
		if err != nil {
			return err
		}

		src, err := safeJoin(dir, path)
		if err != nil {
			return err
		}

		err = os.MkdirAll(filepath.Dir(dest), 0o700)
		if err != nil {
			return &IOError{Path: cachePath, Cause: err}
		}

		err = copyFile(src, dest)
		if err != nil {
			return err
		}
	}

	if err := os.Rename(tmp, cachePath); err != nil {
		// some other worker already put cachePath
		// this is the multiple worker guarantee for the future
		if _, sErr := os.Stat(cachePath); sErr == nil {
			return nil
		}
		return &IOError{Path: cachePath, Cause: err}
	}

	return nil
}

func (c *L1) Get(key, dir string) (bool, error) {
	cachePath := filepath.Join(c.root, key)
	if _, err := os.Stat(cachePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, &IOError{Path: cachePath, Cause: err}
	}

	err := filepath.WalkDir(cachePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return &IOError{Path: path, Cause: err}
		}

		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(cachePath, path)
		if err != nil {
			return &IOError{Path: path, Cause: err}
		}

		dest, err := safeJoin(dir, rel)
		if err != nil {
			return err
		}

		err = os.MkdirAll(filepath.Dir(dest), 0o700)
		if err != nil {
			return &IOError{Path: filepath.Dir(dest), Cause: err}
		}

		err = copyFile(path, dest)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}
