/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

type ManyParser[T any] struct {
	callback func(string, ...T) T
	subParser, sepParser Parser[T]
}

func NewManyParser[T any](callback func(string, ...T) T, subparser Parser[T], sepparser Parser[T]) *ManyParser[T] {
	return &ManyParser[T]{callback: callback, subParser: subparser, sepParser: sepparser}
}

func (p *ManyParser[T]) Set(embedded Parser[T], separator Parser[T]) {
	p.subParser = embedded
	p.sepParser = separator
}

func (p *ManyParser[T]) Match(s *Scanner[T]) (Node[T], bool) {
	var nodes []T
	start := s.position

	i := 0
	lastValidPos := s.position

	for {
		if i > 0 && p.sepParser != nil {
			_, ok := s.applyRule(p.sepParser)
			if !ok {
				break
			}
		}
		i++

		node, ok := s.applyRule(p.subParser)
		if !ok {
			break
		}

		nodes = append(nodes, node.Payload)
		lastValidPos = s.position
	}
	s.setPosition(lastValidPos)

	if len(nodes) >= 1 {
		return Node[T]{Parser: p, Payload: p.callback(s.input[start:s.position], nodes...)}, true
	}

	return Node[T]{}, false
}
