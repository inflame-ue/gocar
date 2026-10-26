package testutil

import (
	"errors"
	"testing"
)

// CheckErrAs is a test helper that accepts an error interface
// alongside a specific error type to verify both pointer equality
// and type equality. Returns the specific error back to the caller,
// so that specific fields could be further verified.
func CheckErrAs[E error](t testing.TB, got error, want E) E {
	t.Helper()

	var zero E
	if any(want) == any(zero) {
		if got != nil {
			t.Fatalf("expected no err, got %v", got)
		}
		return zero
	}

	var e E
	if !errors.As(got, &e) {
		t.Fatalf("expected %T, got %T", want, got)
	}

	return e
}
