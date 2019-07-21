/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "strings"

type AndParser struct {
	subParser []Parser
}

func NewAndParser(subparser ...Parser) *AndParser {
	return &AndParser{subParser: subparser}
}

func (p *AndParser) Set(embedded ...Parser) {
	p.subParser = embedded
}

func (p *AndParser) Description(stack map[Parser]bool) string {
	b := strings.Builder{}
	b.WriteString("And(")
	b.WriteString(writeDebug(p, stack, p.subParser...))
	b.WriteString(")")
	return b.String()
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
