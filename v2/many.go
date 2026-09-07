/*
	(c) 2019-2026 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "sync"

type ManyParser[T any] struct {
	callback             func(string, ...T) T
	subParser, sepParser Parser[T]
	buf                  sync.Pool
	NoMemo               bool

	// Accumulation form - see KleeneParser. Step folds each matched item into an
	// accumulator seeded by Init and mapped by Finish. As with the collecting
	// form the parser still fails when nothing matched.
	Init   func() T
	Step   func(acc, item T) T
	Finish func(acc T) T
}

func NewManyParser[T any](callback func(string, ...T) T, subparser Parser[T], sepparser Parser[T]) *ManyParser[T] {
	p := &ManyParser[T]{callback: callback, subParser: subparser, sepParser: sepparser}
	p.buf.New = func() any {
		buffer := make([]T, 0, 8)
		return &buffer
	}
	return p
}

func (p *ManyParser[T]) Set(embedded Parser[T], separator Parser[T]) {
	p.subParser = embedded
	p.sepParser = separator
}

// SetAccumulator installs the accumulation form (Init / Step / Finish).
func (p *ManyParser[T]) SetAccumulator(init func() T, step func(acc, item T) T, finish func(acc T) T) {
	p.Init = init
	p.Step = step
	p.Finish = finish
}

func (p *ManyParser[T]) Match(s *Scanner[T]) (Node[T], bool) {
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
	lastValidPos := s.position
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
		lastValidPos = s.position
	}
	s.setPosition(lastValidPos)

	if len(nodes) >= 1 {
		return Node[T]{Payload: p.callback(s.input[start:s.position], nodes...)}, true
	}

	return Node[T]{}, false
}

func (p *ManyParser[T]) matchAccumulate(s *Scanner[T]) (Node[T], bool) {
	var acc T
	if p.Init != nil {
		acc = p.Init()
	}
	applyFn := s.applyRule
	if p.NoMemo {
		applyFn = func(rule Parser[T]) (Node[T], bool) { return rule.Match(s) }
	}

	matched := 0
	lastValidPos := s.position
	for {
		if matched > 0 && p.sepParser != nil {
			if _, ok := applyFn(p.sepParser); !ok {
				break
			}
		}

		node, ok := applyFn(p.subParser)
		if !ok {
			break
		}

		acc = p.Step(acc, node.Payload)
		matched++
		lastValidPos = s.position
	}
	s.setPosition(lastValidPos)

	if matched == 0 {
		return Node[T]{}, false
	}
	if p.Finish != nil {
		acc = p.Finish(acc)
	}
	return Node[T]{Payload: acc}, true
}
