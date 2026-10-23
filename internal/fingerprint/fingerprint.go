package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"slices"

	"github.com/inflame-ue/gocar/internal/glob"
	"github.com/inflame-ue/gocar/internal/task"
)

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

		var inputs []string
		for _, pattern := range t.Inputs {
			matches, err := glob.Expand(fp.Dir, pattern)
			if err != nil {
				return nil, &InputError{
					Task: name, 
					Path: pattern,
					Cause: err,
				}
			}
			inputs = append(inputs, matches...)
		}
		
		slices.Sort(inputs)
		for _, path := range slices.Compact(inputs) {
			writeField(h, []byte(path))

			if fileHash, ok := fileHashes[path]; ok {
				writeField(h, fileHash)
				continue
			}

			fileHash, err := hashFile(fp.Dir, path)
			if err != nil {
				return nil, &InputError{
					Task:  name,
					Path:  path,
					Cause: err,
				}
			}
			fileHashes[path] = fileHash
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
