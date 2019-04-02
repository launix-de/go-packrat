package packrat

type EmptyParser struct {
	subParser Parser
	skipWs    bool
}

func NewEmptyParser(subparser Parser) *EmptyParser {
	return &EmptyParser{subParser: subparser}
}

// Set updates the embedded parser
func (p *EmptyParser) Set(subParser Parser) {
	p.subParser = subParser
}

// Match matches only the given string. If skipWs is set to true, leading whitespace according to the scanner's skip regexp is skipped, but not matched by the parser.
func (p *EmptyParser) Match(os *Scanner) (*Scanner, Node) {
	s := os
	if p.skipWs {
		s = s.Copy()
		s.Skip()
	}

	startPosition := s.position
	cached, wasCached := s.memoization[startPosition][p]
	if wasCached {
		nss, node := cached.Scanner, cached.Node
		return nss, node
	}

	var node Node
	subCached, subWasCached := s.memoization[startPosition][p.subParser]
	if subWasCached {
		node = subCached.Node
		s = subCached.Scanner
	} else {
		s, node = p.Match(s)
		s.memoization[startPosition][p.subParser] = scannerNode{Scanner: s, Node: node}
	}

	r := scannerNode{Scanner: s, Node: Node{Matched: node.Matched, Children: []Node{node}, Parser: p}}
	os.memoization[startPosition][p] = r
	if s != nil {
		return r.Scanner, r.Node
	}

	return nil, Node{}
}
