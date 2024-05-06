/*
	(c) 2019, 2023 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge
	Author: Carl-Philip Hänsch

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import (
	"sync"
	"regexp"
	"unicode"
)

type MemoEntry[T any] struct {
	Lr  *Lr[T]
	Ans Node[T]
	Ok bool

	Position int
}

type Head[T any] struct {
	rule        Parser[T]
	involvedSet map[Parser[T]]bool
	evalSet     map[Parser[T]]bool
}

func (s *Scanner[T]) NewHead(rule Parser[T]) *Head[T] {
	h := s.headpool.Get().(*Head[T])
	h.rule = rule
	clear(h.involvedSet)
	clear(h.evalSet)
	return h
}

func (h *Head[T]) IsInvolved(rule Parser[T]) bool {
	if rule == h.rule {
		return true
	}

	_, isInvoled := h.involvedSet[rule]
	return isInvoled
}

func (h *Head[T]) IsEvaluated(rule Parser[T]) bool {
	_, isEvaluated := h.evalSet[rule]
	return isEvaluated
}

type Lr[T any] struct {
	seed Node[T]
	seedOk bool
	rule Parser[T]
	head *Head[T]
}

type Scanner[T any] struct {
	input           string
	remainingInput  string
	position        int
	memoization     map[int]map[Parser[T]]*MemoEntry[T]
	heads           map[int]*Head[T]
	invocationStack [200]Lr[T]
	invocationStackIdx int
	breaks          map[int]bool
	
	headpool        sync.Pool


	skipRegex *regexp.Regexp
}

// Copy clones the scanner state. Memoization and break maps are shared
func (s *Scanner[T]) Copy() *Scanner[T] {
	ns := *s
	return &ns
}

func (s *Scanner[T]) Recall(rule Parser[T], pos int) *MemoEntry[T] {
	mmap, mmapExists := s.memoization[pos]
	var m *MemoEntry[T]
	if mmapExists {
		m = mmap[rule]
	}

	// If not growing a seed parse, just return what is stored in the memo table
	head, headExists := s.heads[pos]
	if !headExists {
		return m
	}

	// Do not evaluate any rule that is not involved in this left recursion
	if m == nil && !head.IsInvolved(rule) {
		return &MemoEntry[T]{Position: s.position}
	}

	// Allow involved rules to be evaluated, but only once, during a seed-growing iteration
	if head.IsEvaluated(rule) {
		delete(head.evalSet, rule)
		node, ok := rule.Match(s)
		return &MemoEntry[T]{Position: s.position, Ans: node, Ok: ok}
	}

	return m
}

func (s *Scanner[T]) SetupLr(rule Parser[T], l *Lr[T]) {
	if l.head == nil {
		l.head = s.NewHead(rule)
	}
	i := s.invocationStackIdx - 1
	for i >= 0 && s.invocationStack[i].head != l.head {
		s.invocationStack[i].head = l.head
		newInvolved := make(map[Parser[T]]bool)
		for k := range l.head.involvedSet {
			newInvolved[k] = true
		}
		newInvolved[s.invocationStack[i].rule] = true
		l.head.involvedSet = newInvolved
		i--
	}
}

func (s *Scanner[T]) GrowLr(rule Parser[T], p int, m *MemoEntry[T], h *Head[T]) (Node[T], bool) {
	s.heads[p] = h
	for {
		s.setPosition(p)
		h.evalSet = make(map[Parser[T]]bool)
		for k, v := range h.involvedSet {
			h.evalSet[k] = v
		}
		ans, ok := rule.Match(s)
		if !ok || s.position <= m.Position {
			break
		}
		m.Lr = nil
		m.Ans = ans
		m.Ok = ok
		m.Position = s.position
	}
	s.headpool.Put(s.heads[p])
	delete(s.heads, p)
	s.setPosition(m.Position)
	return m.Ans, m.Ok
}

func (s *Scanner[T]) LrAnswer(rule Parser[T], pos int, m *MemoEntry[T]) (Node[T], bool) {
	h := m.Lr.head
	if h.rule != rule {
		return m.Lr.seed, m.Lr.seedOk
	}
	m.Ans = m.Lr.seed
	m.Ok = m.Lr.seedOk
	m.Lr = nil

	if !m.Ok {
		return Node[T]{}, false
	}
	return s.GrowLr(rule, pos, m, h)
}

var SkipWhitespaceRegex = regexp.MustCompile("^[\r\n\t ]+")
var SkipWhitespaceAndCommentsRegex = regexp.MustCompile("^(?:/\\*.*?\\*/|[\r\n\t ]+)+") // regex for comments

// skipper: use nil, SkipWhitespaceRegex or your very own regex
func NewScanner[T any](input string, skipper *regexp.Regexp) *Scanner[T] {
	s := &Scanner[T]{input: input, position: 0, memoization: make(map[int]map[Parser[T]]*MemoEntry[T]),
		heads: make(map[int]*Head[T])}
	s.headpool.New = func() any {
		return &Head[T]{nil, make(map[Parser[T]]bool), make(map[Parser[T]]bool)}
	}
	s.remainingInput = s.input
	s.skipRegex = skipper
	s.breaks = make(map[int]bool)

	previousWord := false
	for pos, r := range input {
		currentWord := unicode.In(r, unicode.N, unicode.L, unicode.Pc)
		if !currentWord || !previousWord {
			s.breaks[pos] = true
		}

		previousWord = currentWord
	}

	s.breaks[len(input)] = true

	return s
}

func (s *Scanner[T]) isAtBreak() bool {
	return s.breaks[s.position]
}

func (s *Scanner[T]) move(n int) {
	s.position += n
	s.remainingInput = s.input[s.position:]
}

func (s *Scanner[T]) setPosition(n int) {
	s.position = n
	s.remainingInput = s.input[s.position:]
}

func (s *Scanner[T]) MatchRegexp(r *regexp.Regexp) *string {
	matched := r.FindStringSubmatch(s.remainingInput)
	if matched != nil {
		s.move(len(matched[0]))
		return &matched[0]
	}

	return nil
}

func (s *Scanner[T]) Skip() {
	if s.skipRegex != nil {
		s.MatchRegexp(s.skipRegex)
	}
}
