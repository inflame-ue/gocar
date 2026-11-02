package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"slices"

	"github.com/inflame-ue/gocar/internal/glob"
	"github.com/inflame-ue/gocar/internal/task"
)

type Fingerprinter struct {
	Dir string

	// Reconstruct Fingerprinter to keep the file hashes scoped to per-build basis.
	// Allows to see changes between files in-between builds.
	fileHashes map[string][]byte
}

func NewFingerprinter(dir string) *Fingerprinter {
	return &Fingerprinter{
		Dir:        dir,
	}
}

// Key computes the fingerprint for a single task. The depKeys must contain
// all keys of every task in t.Deps, otherwise Key will return a DependencyError.
func (fp *Fingerprinter) Key(name string, t task.Task, depKeys map[string]string) (string, error) {
	h := sha256.New()
	fp.fileHashes = make(map[string][]byte)

	writeField(h, []byte(t.Cmd))

	for _, output := range canonicalPaths(t.Outputs) {
		writeField(h, []byte(output))
	}

	var inputs []string
	for _, pattern := range t.Inputs {
		matches, err := glob.Expand(fp.Dir, pattern)
		if err != nil {
			return "", &InputError{
				Task:  name,
				Path:  pattern,
				Cause: err,
			}
		}
		inputs = append(inputs, matches...)
	}

	slices.Sort(inputs)
	for _, path := range slices.Compact(inputs) {
		writeField(h, []byte(path))

		if fileHash, ok := fp.fileHashes[path]; ok {
			writeField(h, fileHash)
			continue
		}

		fileHash, err := hashFile(fp.Dir, path)
		if err != nil {
			return "", &InputError{
				Task:  name,
				Path:  path,
				Cause: err,
			}
		}

		fp.fileHashes[path] = fileHash
		writeField(h, fileHash)
	}

	deps := slices.Clone(t.Deps)
	slices.Sort(deps)
	for _, dep := range slices.Compact(deps) {
		depKey, ok := depKeys[dep]
		if !ok {
			return "", &DependencyError{
				Task: name,
				Dep:  dep,
			}
		}
		writeField(h, []byte(depKey))
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func canonicalPaths(paths []string) []string {
	canonical := make([]string, 0, len(paths))
	for _, path := range paths {
		canonical = append(canonical, filepath.ToSlash(filepath.Clean(path)))
	}

	slices.Sort(canonical)
	return slices.Compact(canonical)
}
