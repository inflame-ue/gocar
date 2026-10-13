package glob

import "path/filepath"

func segMatch(pattern, seg string) (bool, error) {
	// I don't like the fact that filepath.Match forces the error
	// return here. Could be resolved by writing from scratch, but
	// no complexity/information would be gained from that.
	return filepath.Match(pattern, seg)
}

func match(pattern, path []string) (bool, error) {
	if len(pattern) == 0 {
		return len(path) == 0, nil
	}

	if len(path) == 0 {
		return false, nil
	}

	segMatched, err := segMatch(pattern[0], path[0])
	if err != nil {
		return false, err
	}

	matched, err := match(pattern[1:], path[1:])
	if err != nil {
		return false, err
	}

	return segMatched && matched, nil
}
