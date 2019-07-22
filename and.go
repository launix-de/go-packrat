/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "strings"

// AndParser accepts an input if all sub parsers accept the input sequentially
type AndParser struct {
	subParser []Parser
}

// NewAndParser constructs a new AndParser with the given sub parsers. An AndParser accepts an input if all sub parsers accept the input sequentially.
func NewAndParser(subparser ...Parser) *AndParser {
	return &AndParser{subParser: subparser}
}

// Set updates the sub parsers. This can be used to construct recursive parsers.
func (p *AndParser) Set(embedded ...Parser) {
	p.subParser = embedded
}

// Match matches all given parsers sequentially.
func (p *AndParser) Match(s *Scanner) *Node {
	var nodes []*Node

	startPosition := s.position
	for _, c := range p.subParser {
		node := s.applyRule(c)
		if node == nil {
			s.setPosition(startPosition)
			return nil
		}
		nodes = append(nodes, node)
	}

	b := strings.Builder{}
	for _, n := range nodes {
		b.WriteString(n.Matched)
	}
	matched := b.String()

	return &Node{Matched: matched, Parser: p, Children: nodes}
}
