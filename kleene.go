/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

type KleeneParser[T any] struct {
	callback func(string, ...T) T
	subParser, sepParser Parser[T]
}

func NewKleeneParser[T any](callback func(string, ...T) T, subparser Parser[T], sepparser Parser[T]) *KleeneParser[T] {
	return &KleeneParser[T]{callback: callback, subParser: subparser, sepParser: sepparser}
}

func (p *KleeneParser[T]) Set(embedded Parser[T], separator Parser[T]) {
	p.subParser = embedded
	p.sepParser = separator
}

// Match matches the embedded parser or the empty string.
func (p *KleeneParser[T]) Match(s *Scanner[T]) (Node[T], bool) {
	var nodes []T
	start := s.position

	i := 0
	lastValidPosition := s.position
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
		if i == 1 {
			start = node.Start
		}

		nodes = append(nodes, node.Payload)
		lastValidPosition = s.position
	}
	s.setPosition(lastValidPosition)

	if len(nodes) == 0 {
		return Node[T]{Start: s.position, Parser: p, Payload: p.callback("")}, true
	}
	return Node[T]{Start: start, Parser: p, Payload: p.callback(s.input[start:s.position], nodes...)}, true
}
