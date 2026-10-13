package glob

import "fmt"

type NoMatchError struct {
	Pattern string
}

func (nme *NoMatchError) Error() string {
	return fmt.Sprintf("no match for pattern: %s", nme.Pattern)
}
