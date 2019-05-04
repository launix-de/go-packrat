/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "regexp"

type RegexParser struct {
	regex  *regexp.Regexp
	skipWs bool
}

func NewRegexParser(rs string, skipWs bool) *RegexParser {
	r := regexp.MustCompile("^" + rs)
	return &RegexParser{regex: r, skipWs: skipWs}
}

// Regex matches only the given regexp. If skipWs is set to true, leading whitespace according to the scanner's skip regexp is skipped, but not matched by the parser.
// Regex panics if rs is not a valid regex string.
func (p *RegexParser) Match(os *Scanner) (*Scanner, Node) {
	s := os.Copy()
	opos := s.position

	if opos >= len(s.memoization) {
		return nil, Node{}
	}
	cached, wasCached := s.memoization[opos][p]
	if wasCached {
		return cached.Scanner, cached.Node
	}

	if p.skipWs {
		s.Skip()
		if !s.isAtBreak() {
			s.memoization[opos][p] = scannerNode{}
			return nil, Node{}
		}
	}


	matched := s.MatchRegexp(p.regex)
	if matched == nil {
		s.memoization[s.position][p] = scannerNode{}
		return nil, Node{}
	}

	if p.skipWs {
		if !s.isAtBreak() {
			s.memoization[opos][p] = scannerNode{}
			return nil, Node{}
		}
	}

	r := scannerNode{Scanner: s, Node: Node{Matched: *matched, Parser: p}}
	s.memoization[opos][p] = r
	return r.Scanner, r.Node
}
