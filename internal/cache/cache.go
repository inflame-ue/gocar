package cache

import (
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
	defer os.RemoveAll(tmp)

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
