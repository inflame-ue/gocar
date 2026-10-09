package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/inflame-ue/gocar/internal/task"
)

func Keys(spec map[string]task.Task, order []string) (map[string]string, error) {
	var digests map[string][]byte

	h := sha256.New()
	for _, name := range order {
		t, ok := spec[name]
		if !ok {
			// placeholder for now
			return nil, errors.New("task name no in spec")
		}

		writeField(h, []byte(t.Cmd))

		digests[name] = h.Sum(nil)
	}

	var keys map[string]string
	for key, digest := range digests {
		keys[key] = hex.EncodeToString(digest)
	}

	return keys, nil
}
