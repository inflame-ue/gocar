package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"slices"

	"github.com/inflame-ue/gocar/internal/task"
)

func normalizedPath(path string) string {
	return ""
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

		for _, input := range t.Inputs {
			filename := normalizedPath(input)
			writeField(h, []byte(filename))

			fileHash, err := contentHash(filename)
			if err != nil {
				return nil, errors.New("failed to hash file content")
			}
			
			writeField(h, fileHash)
		}

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
