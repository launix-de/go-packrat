/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "strings"

type KleeneParser struct {
	subParser, sepParser Parser
}

func NewKleeneParser(subparser Parser, sepparser Parser) *KleeneParser {
	return &KleeneParser{subParser: subparser, sepParser: sepparser}
}

func (p *KleeneParser) Set(embedded Parser, separator Parser) {
	p.subParser = embedded
	p.sepParser = separator
}

func (p *KleeneParser) Description(stack map[Parser]bool) string {
	b := strings.Builder{}
	b.WriteString("Kleene(")
	b.WriteString(writeDebug(p, stack, p.subParser, p.sepParser))
	b.WriteString(")")
	return b.String()
}

// Match matches the embedded parser or the empty string.
func (p *KleeneParser) Match(s *Scanner) *Node {
	var nodes []*Node

	i := 0
	lastValidPosition := s.position
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
		lastValidPosition = s.position
	}
	s.setPosition(lastValidPosition)

	b := strings.Builder{}
	for _, n := range nodes {
		b.WriteString(n.Matched)
	}
	matched := b.String()

	return &Node{Matched: matched, Parser: p, Children: nodes}
}
