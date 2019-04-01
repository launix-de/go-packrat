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

			if nss == nil {
				s.memoization[startPosition][p] = scannerNode{}
				return nil, Node{}
			}
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
