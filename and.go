package packrat

type AndParser struct {
	subParser []Parser
}

func NewAndParser(subparser ...Parser) *AndParser {
	return &AndParser{subParser: subparser}
}

func (p *AndParser) Set(embedded ...Parser) {
	p.subParser = embedded
}

// Match matches all given parsers sequentially.
func (p *AndParser) Match(s *Scanner) (*Scanner, Node) {
	startPosition := s.position
	cached, wasCached := s.memoization[startPosition][p]
	if wasCached {
		nss, node := cached.Scanner, cached.Node
		return nss, node
	}

	ns := s.Copy()
	var nodes []Node

	for _, c := range p.subParser {
		if ns.position >= len(s.memoization) {
			s.memoization[startPosition][p] = scannerNode{}
			return nil, Node{}
		}
		cached, wasCached := s.memoization[ns.position][c]

		var (
			nss  *Scanner
			node Node
		)
		if wasCached {
			nss, node = cached.Scanner, cached.Node
		} else {
			nss, node = c.Match(ns)
			s.memoization[ns.position][c] = scannerNode{Scanner: nss, Node: node}
		}
		if nss == nil {
			s.memoization[startPosition][p] = scannerNode{}
			return nil, Node{}
		}

		ns = nss
		nodes = append(nodes, node)
	}

	endPosition := ns.position
	matched := ns.input[startPosition:endPosition]

	r := scannerNode{Scanner: ns, Node: Node{Matched: matched, Parser: p, Children: nodes}}
	s.memoization[startPosition][p] = r

	return r.Scanner, r.Node
}
