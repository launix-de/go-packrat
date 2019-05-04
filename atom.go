/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "regexp"

type AtomParser struct {
	str    string
	skipWs bool

	regex *regexp.Regexp
}

func NewAtomParser(str string, skipWs bool) *AtomParser {
	p := &AtomParser{str: str, skipWs: skipWs}
	return p
}

// Match matches only the given string. If skipWs is set to true, leading whitespace according to the scanner's skip regexp is skipped, but not matched by the parser.
func (p *AtomParser) Match(os *Scanner) (*Scanner, Node) {
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

	matched := s.MatchString(p.str)
	if matched == nil {
		s.memoization[opos][p] = scannerNode{}
		return nil, Node{}
	}

	if p.skipWs {
		if !s.isAtBreak() {
			s.memoization[opos][p] = scannerNode{}
			return nil, Node{}
		}
	}

	r := scannerNode{Scanner: s, Node: Node{Matched: p.str, Parser: p}}
	s.memoization[opos][p] = r
	return r.Scanner, r.Node
}
