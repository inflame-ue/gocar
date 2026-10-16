package glob

import "path/filepath"

const (
	unknown int8 = 0
	success int8 = 1
	fail    int8 = -1
)

func clasiffyMatch(result bool) int8 {
	if !result {
		return fail
	}
	return success
}

type matcher struct {
	pattern, path []string
	memo          [][]int8
}

func (m *matcher) recurse(pi, qi int) (bool, error) {
	if pi == len(m.pattern) {
		match := qi == len(m.path)
		m.memo[pi][qi] = clasiffyMatch(match)
		return match, nil
	}

	// if success, then no need to go through recursion branches
	if m.memo[pi][qi] != 0 {
		return m.memo[pi][qi] == 1, nil
	}

	if m.pattern[pi] == "**" {
		if match, err := m.recurse(pi+1, qi); err != nil || match {
			m.memo[pi+1][qi] = clasiffyMatch(match)
			return match, err
		}

		if qi < len(m.path) {
			match, err := m.recurse(pi, qi+1)
			m.memo[pi][qi+1] = clasiffyMatch(match)
			return match, err
		}

		return false, nil
	}

	if qi == len(m.path) {
		m.memo[pi][qi] = fail
		return false, nil
	}

	segMatched, err := segMatch(m.pattern[pi], m.path[qi])
	if err != nil {
		m.memo[pi][qi] = fail
		return false, err
	}

	matched, err := m.recurse(pi+1, qi+1)
	if err != nil {
		m.memo[pi][qi] = fail
		return false, err
	}

	match := segMatched && matched
	m.memo[pi][qi] = clasiffyMatch(match)
	return segMatched && matched, nil
}

func match(pattern, path []string) (bool, error) {
	memo := make([][]int8, len(pattern)+1)
	for i := range memo {
		memo[i] = make([]int8, len(path)+1)
	}

	m := &matcher{
		pattern: pattern,
		path:    path,
		memo:    memo,
	}

	return m.recurse(0, 0)
}

// naiveMatch utilizes naive recursion without memoization
// to match the double star segments. It is not used by gocar
// and is kept for the purposes of a benchmark comparison.
func naiveMatch(pattern, path []string) (bool, error) {
	if len(pattern) == 0 {
		return len(path) == 0, nil
	}

	if pattern[0] == "**" {
		if ok, err := naiveMatch(pattern[1:], path); err != nil || ok {
			return ok, err
		}

		if len(path) > 0 {
			return naiveMatch(pattern, path[1:])
		}

		return false, nil
	}

	if len(path) == 0 {
		return false, nil
	}

	segMatched, err := segMatch(pattern[0], path[0])
	if err != nil {
		return false, err
	}

	matched, err := naiveMatch(pattern[1:], path[1:])
	if err != nil {
		return false, err
	}

	return segMatched && matched, nil
}

// segMatch is a wrapper of filepath.Match that matches a single segment.
func segMatch(pattern, seg string) (bool, error) {
	// I don't like the fact that filepath.Match forces the error
	// return here. Could be resolved by writing from scratch, but
	// no complexity/information would be gained from that.
	return filepath.Match(pattern, seg)
}
