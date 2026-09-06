/*
	(c) 2019-2026 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "sync"

type KleeneParser[T any] struct {
	callback             func(string, ...T) T
	subParser, sepParser Parser[T]
	buf                  sync.Pool
	NoMemo               bool

	// Accumulation form. When Step is non-nil the repeat folds every matched
	// item into an accumulator instead of collecting them for callback: Init
	// seeds the accumulator, Step folds one item into it, Finish maps the final
	// accumulator to the result payload. The accumulator is created fresh on
	// every Match, so an outer backtrack that re-enters this parser simply
	// re-inits - there is nothing to snapshot. Init and Finish are optional.
	Init   func() T
	Step   func(acc, item T) T
	Finish func(acc T) T
}

func NewKleeneParser[T any](callback func(string, ...T) T, subparser Parser[T], sepparser Parser[T]) *KleeneParser[T] {
	p := &KleeneParser[T]{callback: callback, subParser: subparser, sepParser: sepparser}
	p.buf.New = func() any {
		buffer := make([]T, 0, 8)
		return &buffer
	}
	return p
}

func (p *KleeneParser[T]) Set(embedded Parser[T], separator Parser[T]) {
	p.subParser = embedded
	p.sepParser = separator
}

// SetAccumulator installs the accumulation form (Init / Step / Finish).
func (p *KleeneParser[T]) SetAccumulator(init func() T, step func(acc, item T) T, finish func(acc T) T) {
	p.Init = init
	p.Step = step
	p.Finish = finish
}

// Match matches the embedded parser or the empty string.
func (p *KleeneParser[T]) Match(s *Scanner[T]) (Node[T], bool) {
	if p.Step != nil {
		return p.matchAccumulate(s)
	}

	buffer := p.buf.Get().(*[]T)
	nodes := (*buffer)[:0]
	defer func() {
		clear(nodes)
		*buffer = nodes[:0]
		p.buf.Put(buffer)
	}()
	start := s.position

	i := 0
	lastValidPosition := s.position
	applyFn := s.applyRule
	if p.NoMemo {
		applyFn = func(rule Parser[T]) (Node[T], bool) { return rule.Match(s) }
	}
	for {
		if i > 0 && p.sepParser != nil {
			_, ok := applyFn(p.sepParser)
			if !ok {
				break
			}
		}
		i++

		node, ok := applyFn(p.subParser)
		if !ok {
			break
		}

		nodes = append(nodes, node.Payload)
		lastValidPosition = s.position
	}
	s.setPosition(lastValidPosition)

	if len(nodes) == 0 {
		return Node[T]{Payload: p.callback("")}, true
	}
	return Node[T]{p.callback(s.input[start:s.position], nodes...)}, true
}

func (p *KleeneParser[T]) matchAccumulate(s *Scanner[T]) (Node[T], bool) {
	var acc T
	if p.Init != nil {
		acc = p.Init()
	}
	applyFn := s.applyRule
	if p.NoMemo {
		applyFn = func(rule Parser[T]) (Node[T], bool) { return rule.Match(s) }
	}

	i := 0
	lastValidPosition := s.position
	for {
		if i > 0 && p.sepParser != nil {
			if _, ok := applyFn(p.sepParser); !ok {
				break
			}
		}
		i++

		node, ok := applyFn(p.subParser)
		if !ok {
			break
		}

		acc = p.Step(acc, node.Payload)
		lastValidPosition = s.position
	}
	s.setPosition(lastValidPosition)

	if p.Finish != nil {
		acc = p.Finish(acc)
	}
	return Node[T]{Payload: acc}, true
}
