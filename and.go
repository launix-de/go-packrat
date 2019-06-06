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

// Match matches all given parsers sequentially.
func (p *AndParser) Match(s *Scanner) (*Scanner, Node) {
	var nodes []Node

	for _, c := range p.subParser {
		var node Node
		s, node = match(s, c)
		if s == nil {
			return nil, Node{}
		}
		nodes = append(nodes, node)
	}

	b := strings.Builder{}
	for _, n := range nodes {
		b.WriteString(n.Matched)
	}
	matched := b.String()

	r := scannerNode{Scanner: s, Node: Node{Matched: matched, Parser: p, Children: nodes}}
	return r.Scanner, r.Node
}
