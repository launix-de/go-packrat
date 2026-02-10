/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

type OrParser[T any] struct {
	subParser []Parser[T]
	charMap   *[256][]int
}

func NewOrParser[T any](subparser ...Parser[T]) *OrParser[T] {
	return &OrParser[T]{subParser: subparser}
}

func (p *OrParser[T]) Set(embedded ...Parser[T]) {
	p.subParser = embedded
}

// SetCharMap installs a first-byte dispatch table. cm[b] contains the indices
// into subParser that can match when the first non-whitespace byte is b.
// When set, Match skips whitespace and only tries the listed candidates
// instead of all sub-parsers sequentially.
func (p *OrParser[T]) SetCharMap(cm [256][]int) {
	p.charMap = &cm
}

// Match tries sub-parsers until one succeeds. If a charMap is set,
// only the sub-parsers indexed by the first non-whitespace byte are tried.
func (p *OrParser[T]) Match(s *Scanner[T]) (Node[T], bool) {
	if p.charMap != nil {
		s.Skip()
		if s.position >= len(s.input) {
			return Node[T]{}, false
		}
		candidates := p.charMap[s.input[s.position]]
		startPosition := s.position
		for _, idx := range candidates {
			node, ok := s.applyRule(p.subParser[idx])
			if ok {
				return Node[T]{Payload: node.Payload}, true
			}
			s.setPosition(startPosition)
		}
		return Node[T]{}, false
	}

	startPosition := s.position
	for _, c := range p.subParser {
		node, ok := s.applyRule(c)
		if ok {
			return Node[T]{Payload: node.Payload}, true
		}
		s.setPosition(startPosition)
	}

	return Node[T]{}, false
}
