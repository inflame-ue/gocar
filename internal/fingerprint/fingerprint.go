package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"
	"slices"

	"github.com/inflame-ue/gocar/internal/task"
)

func normalizePath(dir, path string) (string, error) {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return "", err
	}

	return filepath.ToSlash(filepath.Clean(rel)), nil
}

type Fingerprinter struct {
	Dir string
}

func (fp *Fingerprinter) Keys(spec map[string]task.Task, order []string) (map[string]string, error) {
	digests := make(map[string][]byte, len(order))

	for _, name := range order {
		h := sha256.New()

		t, ok := spec[name]
		if !ok {
			// placeholder for now
			return nil, errors.New("task name no in spec")
		}

		writeField(h, []byte(t.Cmd))

		for _, path := range t.Inputs {
			normalized, err := normalizePath(fp.Dir, path)
			if err != nil {
				return nil, errors.New("failed to normalize path")
			}
			writeField(h, []byte(normalized))

			fileHash, err := hashContent(fp.Dir, normalized)
			if err != nil {
				return nil, errors.New("failed to hash file content")
			}
			writeField(h, fileHash)
		}

		// Note: Every dependency is already present in the digest
		// because order is topologically sorted.
		deps := slices.Clone(t.Deps)
		slices.Sort(deps)
		for _, dep := range slices.Compact(deps) {
			writeField(h, digests[dep])
		}

		digests[name] = h.Sum(nil)
	}

	keys := make(map[string]string, len(order))
	for key, digest := range digests {
		keys[key] = hex.EncodeToString(digest)
	}

	return keys, nil
}
