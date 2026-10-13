package glob

import (
	"errors"
	"slices"
	"testing"
)

func checkPatternErr(t *testing.T, got error, want *PatternError) {
	t.Helper()

	if got == nil && want == nil {
		return
	}

	if got != nil && want == nil {
		t.Fatalf("expected no err, got %v", got)
	}

	if got == nil && want != nil {
		t.Fatal("expected an err, got no err instead")
	}

	var pe *PatternError
	if !errors.As(got, &pe) {
		t.Fatalf("expected err type %T, got %T instead", want, got)
	}

	if pe.Pattern != want.Pattern {
		t.Errorf("expected pattern %s, got %s instead", want.Pattern, pe.Pattern)
	}
}

func TestSplitSegments(t *testing.T) {
	tests := map[string]struct {
		pattern      string
		wantSegments []string
		wantErr      *PatternError
	}{
		"valid pattern": {
			pattern:      "src/**/*.go",
			wantSegments: []string{"src", "**", "*.go"},
			wantErr:      nil,
		},
		"absolute path": {
			pattern:      "/home/user/src/**/*.go",
			wantSegments: nil,
			wantErr: &PatternError{
				Pattern: "/home/user/src/**/*.go",
			},
		},
		"double star inside segment": {
			pattern:      "s**rc/**/*.go",
			wantSegments: nil,
			wantErr: &PatternError{
				Pattern: "s**rc/**/*.go",
			},
		},
		"escaping pattern": {
			pattern:      "../*.go",
			wantSegments: nil,
			wantErr: &PatternError{
				Pattern: "../*.go",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			segments, err := splitSegments(tc.pattern)
			checkPatternErr(t, err, tc.wantErr)

			if !slices.Equal(segments, tc.wantSegments) {
				t.Errorf("expected segments %v, got %v instead", tc.wantSegments, segments)
			}
		})
	}
}
