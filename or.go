/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

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
func (p *OrParser) Match(s *Scanner) *Node {
	startPosition := s.position
	for _, c := range p.subParser {
		node := s.applyRule(c)
		if node != nil {
			return &Node{Matched: node.Matched, Start: node.Start, Parser: p, Children: []*Node{node}}
		}
		s.setPosition(startPosition)
	}

	return nil
}
