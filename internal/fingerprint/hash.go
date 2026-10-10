package fingerprint

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"io"
	"os"
	"path/filepath"
)

func writeField(h hash.Hash, b []byte) {
	var n [8]byte
	binary.BigEndian.PutUint64(n[:], uint64(len(b)))
	h.Write(n[:])
	h.Write(b)
}

func hashFile(dir, filename string) ([]byte, error) {
	path := filepath.Join(dir, filepath.FromSlash(filename))

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileHash := sha256.New()
	io.Copy(fileHash, file)
	return fileHash.Sum(nil), nil
}
