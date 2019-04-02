/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

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
	startPosition := s.position
	cached, wasCached := s.memoization[startPosition][p]
	if wasCached {
		nss, node := cached.Scanner, cached.Node
		return nss, node
	}

	ns := s.Copy()
	var (
		nss  *Scanner
		node Node
	)

	if ns.position >= len(s.memoization) {
		return nil, Node{}
	}
	subCached, subWasCached := ns.memoization[ns.position][p.subParser]
	if subWasCached {
		nss, node = subCached.Scanner, subCached.Node
	} else {
		nss, node = p.subParser.Match(ns)
		ns.memoization[ns.position][p.subParser] = scannerNode{Scanner: nss, Node: node}
	}

	var r scannerNode
	if nss == nil {
		r = scannerNode{Scanner: ns, Node: Node{Matched: emptyString, Parser: p}}
	} else {
		r = scannerNode{Scanner: ns, Node: Node{Matched: node.Matched, Parser: p, Children: []Node{node}}}
	}

	s.memoization[startPosition][p] = r
	return r.Scanner, r.Node
}
