/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "regexp"

type AtomParser struct {
	r    *regexp.Regexp
	skipWs bool
}

func NewAtomParser(str string, caseInsensitive bool, skipWs bool) *AtomParser {
	prefix := ""
	if caseInsensitive {
		prefix += "(?i)"
	}
	prefix += "^"
	r := regexp.MustCompile(prefix + regexp.QuoteMeta(str))
	p := &AtomParser{skipWs: skipWs, r: r}	
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

	matched := s.MatchRegexp(p.r)
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

	r := scannerNode{Scanner: s, Node: Node{Matched: *matched, Parser: p}}
	s.memoization[opos][p] = r
	return r.Scanner, r.Node
}
