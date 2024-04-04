/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

type OrParser[T any] struct {
	subParser []Parser[T]
}

func NewOrParser[T any](subparser ...Parser[T]) *OrParser[T] {
	return &OrParser[T]{subParser: subparser}
}

func (p *OrParser[T]) Set(embedded ...Parser[T]) {
	p.subParser = embedded
}

// Match matches all given parsers sequentially.
func (p *OrParser[T]) Match(s *Scanner[T]) (Node[T], bool) {
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
