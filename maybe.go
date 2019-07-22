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
func (p *MaybeParser) Match(s *Scanner) *Node {
	startPosition := s.position
	node := s.applyRule(p.subParser)

	if node == nil {
		s.setPosition(startPosition)
		return &Node{Matched: emptyString, Parser: p}
	}

	return &Node{Matched: node.Matched, Parser: p, Children: []*Node{node}}
}
