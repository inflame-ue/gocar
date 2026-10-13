package glob

import "fmt"

type NoMatchError struct {
	Pattern string
}

func (nme *NoMatchError) Error() string {
	return fmt.Sprintf("no match for pattern: %s", nme.Pattern)
}

type PatternError struct {
	Pattern, Msg string
}

func (pe *PatternError) Error() string {
	return fmt.Sprintf("malformed pattern %s: %s",  pe.Pattern, pe.Msg)
}
