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
	p := filepath.FromSlash(path)

	if !filepath.IsAbs(p) {
		p = filepath.Join(dir, p)
	}

	rel, err := filepath.Rel(dir, p)
	if err != nil {
		return "", err
	}

	return filepath.ToSlash(filepath.Clean(rel)), nil
}

type Fingerprinter struct {
	Dir string
}

func NewFingerprinter(dir string) *Fingerprinter {
	return &Fingerprinter{Dir: dir}
}

func (fp *Fingerprinter) Keys(spec map[string]task.Task, order []string) (map[string]string, error) {
	digests := make(map[string][]byte, len(order))
	fileHashes := make(map[string][]byte)

	for _, name := range order {
		h := sha256.New()

		t, ok := spec[name]
		if !ok {
			return nil, errors.New("task not in spec")
		}
		writeField(h, []byte(t.Cmd))

		inputs := slices.Clone(t.Inputs)
		slices.Sort(inputs)
		for _, path := range slices.Compact(inputs) {
			normalized, err := normalizePath(fp.Dir, path)
			if err != nil {
				return nil, &InputError{
					Task:  name,
					Path:  path,
					Cause: err,
				}
			}
			writeField(h, []byte(normalized))

			if fileHash, ok := fileHashes[normalized]; ok {
				writeField(h, fileHash)
				continue
			}

			fileHash, err := hashFile(fp.Dir, normalized)
			if err != nil {
				return nil, &InputError{
					Task:  name,
					Path:  path,
					Cause: err,
				}
			}
			fileHashes[normalized] = fileHash
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
