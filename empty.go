/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

type EmptyParser struct {
	// Stub field to prevent compiler from optimizing out &EmptyParser{}
	_hidden bool
}

func NewEmptyParser() *EmptyParser {
	return &EmptyParser{}
}

// Match matches only the given string. If skipWs is set to true, leading whitespace according to the scanner's skip regexp is skipped, but not matched by the parser.
func (p *EmptyParser) Match(os *Scanner) (*Scanner, Node) {
	s := os.Copy()
	startPosition := s.position
	cached, wasCached := s.memoization[startPosition][p]
	if wasCached {
		nss, node := cached.Scanner, cached.Node
		return nss, node
	}

	r := scannerNode{Scanner: s, Node: Node{Matched: emptyString, Parser: p}}
	os.memoization[startPosition][p] = r
	if s != nil {
		return r.Scanner, r.Node
	}

	return nil, Node{}
}
