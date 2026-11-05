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

	// The file hashes are memoized on per-build basis.
	// In other words, when a new Fingerprinter is contstructed,
	// the memoization has to be re-done.
	fileHashes map[string][]byte
}

func NewFingerprinter(dir string) *Fingerprinter {
	return &Fingerprinter{
		Dir:        dir,
		fileHashes: make(map[string][]byte),
	}
}

// Key computes the fingerprint for a single task. The depKeys must contain
// all keys of every task in t.Deps, otherwise Key will return a DependencyError.
func (fp *Fingerprinter) Key(name string, t task.Task, depKeys map[string]string) (string, error) {
	h := sha256.New()

	if fp.fileHashes == nil {
		fp.fileHashes = make(map[string][]byte)
	}

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

// Forget removes the specified literal paths or glob patterns
// from the hash map of memoized file content bytes.
//
// This primarily exists to handle the edge case, when a task
// writes to an output file that another task down the execution graph
// takes on input, which causes a stale memo.
func (fp *Fingerprinter) Forget(patterns ...string) {
	for _, pattern := range patterns {
		delete(fp.fileHashes, canonicalPath(pattern))

		paths, err := glob.Expand(fp.Dir, pattern)
		if err != nil {
			// ignore failed expansions, since it's a cache clean
			// they are not important enough to warrant a build failure
			continue
		}

		for _, path := range paths {
			delete(fp.fileHashes, path)
		}
	}
}

// Clear completely empties the memoized map of file hashes,
// forcing each hash for file content to be re-computed.
func (fp *Fingerprinter) Clear() {
	clear(fp.fileHashes)
}

func canonicalPath(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func canonicalPaths(paths []string) []string {
	canonical := make([]string, 0, len(paths))
	for _, path := range paths {
		canonical = append(canonical, canonicalPath(path))
	}

	slices.Sort(canonical)
	return slices.Compact(canonical)
}
