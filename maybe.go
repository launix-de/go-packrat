package packrat

type MaybeParser struct {
	subParser Parser
}

func NewMaybeParser(subparser Parser) *MaybeParser {
	return &MaybeParser{subParser: subparser}
}

func (p *MaybeParser) Set(embedded Parser) {
	p.subParser = embedded
}

// Match matches the embedded parser or the empty string.
func (p *MaybeParser) Match(s *Scanner) (*Scanner, Node) {
	ns := s.Copy()
	var (
		nss  *Scanner
		node Node
	)

	if ns.position >= len(s.memoization) {
		return nil, Node{}
	}
	cached, wasCached := ns.memoization[ns.position][p.subParser]
	if wasCached {
		nss, node = cached.Scanner, cached.Node
	} else {
		nss, node = p.subParser.Match(ns)
		ns.memoization[ns.position][p.subParser] = scannerNode{Scanner: nss, Node: node}
	}

	if nss == nil {
		return ns, Node{Matched: emptyString, Parser: p}
	}
	return ns, Node{Matched: node.Matched, Parser: p, Children: []Node{node}}
}
