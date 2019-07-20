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
func (p *EmptyParser) Match(s *Scanner) (*Scanner, Node) {
	return s, Node{Matched: emptyString, Parser: p}
}

func (p *EmptyParser) Description(stack map[Parser]bool) string {
	return "Empty"
}
