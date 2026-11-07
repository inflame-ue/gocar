package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile_PermissionRoundTrip(t *testing.T) {
	testDir := t.TempDir()
	src, dst := filepath.Join(testDir, "main.go"), filepath.Join(testDir, "other.go")
	srcPerm := os.FileMode(0o755)

	err := os.WriteFile(src, []byte("package main\n"), srcPerm)
	if err != nil {
		t.Fatalf("write file: %v", err)
	}

	err = copyFile(src, dst)
	if err != nil {
		t.Fatalf("copy file: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}

	dstPerm := info.Mode().Perm()
	if dstPerm != srcPerm {
		t.Errorf("expected dst perm to be %d, got %d instead", srcPerm, dstPerm)
	}
}
