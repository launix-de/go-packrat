/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "strings"

type MaybeParser struct {
	subParser Parser
}

func NewMaybeParser(subparser Parser) *MaybeParser {
	return &MaybeParser{subParser: subparser}
}

func (p *MaybeParser) Set(embedded Parser) {
	p.subParser = embedded
}

func (p *MaybeParser) Description(stack map[Parser]bool) string {
	b := strings.Builder{}
	b.WriteString("Maybe(")
	b.WriteString(writeDebug(p, stack, p.subParser))
	b.WriteString(")")
	return b.String()
}

// Match matches the embedded parser or the empty string.
func (p *MaybeParser) Match(s *Scanner) (*Scanner, Node) {
	ns, node := match(s, p.subParser)

	if ns == nil {
		return s, Node{Matched: emptyString, Parser: p}
	}

	return ns, Node{Matched: node.Matched, Parser: p, Children: []Node{node}}
}
