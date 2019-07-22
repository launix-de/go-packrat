/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "strings"

type ManyParser struct {
	subParser, sepParser Parser
}

func NewManyParser(subparser Parser, sepparser Parser) *ManyParser {
	return &ManyParser{subParser: subparser, sepParser: sepparser}
}

func (p *ManyParser) Set(embedded Parser, separator Parser) {
	p.subParser = embedded
	p.sepParser = separator
}

func (p *ManyParser) Match(s *Scanner) *Node {
	var nodes []*Node

	i := 0
	lastValidPos := s.position

	for {
		matchedsep := false
		var sepnode *Node

		if i > 0 && p.sepParser != nil {
			sepnode = s.applyRule(p.sepParser)
			if sepnode == nil {
				break
			}

			matchedsep = true
		}
		i++

		node := s.applyRule(p.subParser)
		if node == nil {
			break
		}

		if matchedsep {
			nodes = append(nodes, sepnode)
		}

		nodes = append(nodes, node)
		lastValidPos = s.position
	}
	s.setPosition(lastValidPos)

	if len(nodes) >= 1 {
		b := strings.Builder{}
		for _, n := range nodes {
			b.WriteString(n.Matched)
		}
		matched := b.String()

		return &Node{Matched: matched, Parser: p, Children: nodes}
	}

	return nil
}
