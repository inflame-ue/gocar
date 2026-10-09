package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/inflame-ue/gocar/internal/task"
)

type Fingerprinter struct {
	Dir string
}

func (f *Fingerprinter) Keys(spec map[string]task.Task, order []string) (map[string]string, error) {
	digests := make(map[string][]byte, len(order))

	for _, name := range order {
		h := sha256.New()

		t, ok := spec[name]
		if !ok {
			// placeholder for now
			return nil, errors.New("task name no in spec")
		}

		writeField(h, []byte(t.Cmd))
		digests[name] = h.Sum(nil)
	}

	keys := make(map[string]string, len(order))
	for key, digest := range digests {
		keys[key] = hex.EncodeToString(digest)
	}

	return keys, nil
}
