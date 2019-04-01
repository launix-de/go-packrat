package packrat

type OrParser struct {
	subParser []Parser
}

func NewOrParser(subparser ...Parser) *OrParser {
	return &OrParser{subParser: subparser}
}

func (p *OrParser) Set(embedded ...Parser) {
	p.subParser = embedded
}

// Match matches all given parsers sequentially.
func (p *OrParser) Match(s *Scanner) (*Scanner, Node) {
	startposition := s.position
	cached, wasCached := s.memoization[startposition][p]
	if wasCached {
		nss, node := cached.Scanner, cached.Node
		return nss, node
	}

	for _, c := range p.subParser {
		ns := s.Copy()

		if ns.position >= len(s.memoization) {
			return nil, Node{}
		}
		cached, wasCached := ns.memoization[ns.position][c]
		if wasCached {
			nss, node := cached.Scanner, cached.Node
			if nss == nil {
				continue
			}

			r := scannerNode{Scanner: nss, Node: Node{Matched: node.Matched, Parser: p, Children: []Node{node}}}
			s.memoization[s.position][p] = r
			return r.Scanner, r.Node
		}

		nss, node := c.Match(ns)
		s.memoization[ns.position][c] = scannerNode{Scanner: nss, Node: node}

		if nss == nil {
			continue
		}

		r := scannerNode{Scanner: nss, Node: Node{Matched: node.Matched, Parser: p, Children: []Node{node}}}
		s.memoization[startposition][p] = r
		return r.Scanner, r.Node
	}

	return nil, Node{}
}
