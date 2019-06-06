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

func (p *ManyParser) Match(s *Scanner) (*Scanner, Node) {
	var nodes []Node

	i := 0
	for {
		ns := s
		matchedsep := false
		var sepnode Node

		if i > 0 && p.sepParser != nil {
			ns, sepnode = match(s, p.sepParser)
			if ns == nil {
				break
			}

			matchedsep = true
		}
		i++

		var node Node
		ns, node = match(ns, p.subParser)
		if ns == nil {
			break
		}

		if matchedsep {
			nodes = append(nodes, sepnode)
		}

		nodes = append(nodes, node)
		s = ns
	}

	if len(nodes) >= 1 {
		b := strings.Builder{}
		for _, n := range nodes {
			b.WriteString(n.Matched)
		}
		matched := b.String()

		return s, Node{Matched: matched, Parser: p, Children: nodes}
	}

	return nil, Node{}
}
