package glob

import "testing"

func TestMatch(t *testing.T) {
	tests := map[string]struct {
		pattern, path []string
		wantMatch     bool
	}{
		"literal": {
			pattern:   []string{"src", "main.go"},
			path:      []string{"src", "main.go"},
			wantMatch: true,
		},
		"literal no match": {
			pattern:   []string{"src", "main.go"},
			path:      []string{"src", "other.go"},
			wantMatch: false,
		},
		"star": {
			pattern:   []string{"*.go"},
			path:      []string{"main.go"},
			wantMatch: true,
		},
		"star no match": {
			pattern:   []string{"*.go"},
			path:      []string{"main.py"},
			wantMatch: false,
		},
		"double star": {
			pattern: []string{"src", "**", "main.go"},
			path: []string{"src", "cmd", "cli", "main.go"},
			wantMatch: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ok, err := match(tc.pattern, tc.path)

			if err != nil {
				t.Fatalf("match failed: %v", err)
			}

			if ok != tc.wantMatch {
				t.Errorf("expected match %v, got match %v", tc.wantMatch, ok)
			}
		})
	}
}
