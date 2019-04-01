package packrat

type AtomParser struct {
	str    string
	skipWs bool
}

func NewAtomParser(str string, skipWs bool) *AtomParser {
	return &AtomParser{str: str, skipWs: skipWs}
}

// Match matches only the given string. If skipWs is set to true, leading whitespace according to the scanner's skip regexp is skipped, but not matched by the parser.
func (p *AtomParser) Match(os *Scanner) (*Scanner, Node) {
	s := os.Copy()
	if p.skipWs {
		s.Skip()
	}
	opos := s.position

	if opos >= len(s.memoization) {
		return nil, Node{}
	}
	cached, wasCached := s.memoization[opos][p]
	if wasCached {
		return cached.Scanner, cached.Node
	}

	matched := s.MatchString(p.str)
	if matched != nil {
		r := scannerNode{Scanner: s, Node: Node{Matched: p.str, Parser: p}}
		s.memoization[opos][p] = r
		return r.Scanner, r.Node
	}

	s.memoization[opos][p] = scannerNode{}
	return nil, Node{}
}
