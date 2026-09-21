package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTaskSpecification(t *testing.T) {
	filename := filepath.Join("testdata", "smoke.yaml")

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	tasks, err := Parse(data)
	if err != nil {
		t.Error(err)
	}

	t.Logf("parsed: %v", tasks)
}
