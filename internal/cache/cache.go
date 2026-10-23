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
		return &IOError{
			Path:  c.root,
			Cause: err,
		}
	}
	
	cachePath := filepath.Join(c.root, key)
	tmp, err := os.MkdirTemp(c.root, "tmp-*")
	if err != nil {
		return &IOError{
			Path: cachePath,
			Cause: err,
		}
	}
	defer os.RemoveAll(tmp)

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
			return &IOError{
				Path: cachePath, 
				Cause: err,
			}
		}

		err = copyFile(src, dest)
		if err != nil {
			return err
		}
	}

	err = os.Rename(tmp, cachePath)
	if err != nil {
		return &IOError{
			Path:  cachePath,
			Cause: err,
		}
	}

	return nil
}
