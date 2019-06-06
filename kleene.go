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

// Match matches the embedded parser or the empty string.
func (p *KleeneParser) Match(s *Scanner) (*Scanner, Node) {
	var nodes []Node

	i := 0
	for {
		ns := s
		matchedsep := false
		var sepnode Node

		if i > 0 && p.sepParser != nil {
			ns, sepnode = match(ns, p.sepParser)
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

	b := strings.Builder{}
	for _, n := range nodes {
		b.WriteString(n.Matched)
	}
	matched := b.String()
	
	return s, Node{Matched: matched, Parser: p, Children: nodes}
}
