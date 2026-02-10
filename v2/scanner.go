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

	// internal: linked list within a position's memo chain
	rule     Parser[T]
	nextMemo *MemoEntry[T]
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
	next *Lr[T]
}

type Scanner[T any] struct {
	input           string
	remainingInput  string
	position        int
	memoization     []*MemoEntry[T]
	heads           map[int]*Head[T]
	invocationStack *Lr[T]
	breaks          []bool

	headpool        sync.Pool
	lrPool          sync.Pool

	// MemoEntry slab allocator: batch-allocate to reduce per-entry heap allocs
	memoSlab     []MemoEntry[T]
	memoSlabPos  int

	skipRegex *regexp.Regexp
}

const memoSlabSize = 256

func (s *Scanner[T]) newMemoEntry() *MemoEntry[T] {
	if s.memoSlabPos >= len(s.memoSlab) {
		s.memoSlab = make([]MemoEntry[T], memoSlabSize)
		s.memoSlabPos = 0
	}
	m := &s.memoSlab[s.memoSlabPos]
	s.memoSlabPos++
	return m
}

// Copy clones the scanner state. Memoization and break slices are shared.
// Pools are shared by pointer indirection through the original scanner.
func (s *Scanner[T]) Copy() *Scanner[T] {
	ns := &Scanner[T]{
		input:           s.input,
		remainingInput:  s.remainingInput,
		position:        s.position,
		memoization:     s.memoization,
		heads:           s.heads,
		invocationStack: s.invocationStack,
		breaks:          s.breaks,
		memoSlab:        s.memoSlab,
		memoSlabPos:     s.memoSlabPos,
		skipRegex:       s.skipRegex,
	}
	ns.headpool.New = s.headpool.New
	ns.lrPool.New = s.lrPool.New
	return ns
}

func (s *Scanner[T]) memoLookup(pos int, rule Parser[T]) *MemoEntry[T] {
	for m := s.memoization[pos]; m != nil; m = m.nextMemo {
		if m.rule == rule {
			return m
		}
	}
	return nil
}

func (s *Scanner[T]) Recall(rule Parser[T], pos int) *MemoEntry[T] {
	m := s.memoLookup(pos, rule)

	// If not growing a seed parse, just return what is stored in the memo table
	head, headExists := s.heads[pos]
	if !headExists {
		return m
	}

	// Do not evaluate any rule that is not involved in this left recursion
	if m == nil && !head.IsInvolved(rule) {
		me := s.newMemoEntry()
		me.Position = s.position
		return me
	}

	// Allow involved rules to be evaluated, but only once, during a seed-growing iteration
	if head.IsEvaluated(rule) {
		delete(head.evalSet, rule)
		node, ok := rule.Match(s)
		me := s.newMemoEntry()
		me.Position = s.position
		me.Ans = node
		me.Ok = ok
		return me
	}

	return m
}

func (s *Scanner[T]) SetupLr(rule Parser[T], l *Lr[T]) {
	if l.head == nil {
		l.head = s.NewHead(rule)
	}
	stack := s.invocationStack
	for stack != nil && stack.head != l.head {
		stack.head = l.head
		newInvolved := make(map[Parser[T]]bool)
		for k := range l.head.involvedSet {
			newInvolved[k] = true
		}
		newInvolved[stack.rule] = true
		l.head.involvedSet = newInvolved
		stack = stack.next
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
	s := &Scanner[T]{input: input, position: 0, memoization: make([]*MemoEntry[T], len(input)+1),
		heads: make(map[int]*Head[T])}
	s.headpool.New = func() any {
		return &Head[T]{nil, make(map[Parser[T]]bool), make(map[Parser[T]]bool)}
	}
	s.lrPool.New = func() any { return &Lr[T]{} }
	s.remainingInput = s.input
	s.skipRegex = skipper
	s.breaks = make([]bool, len(input)+1)

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
	loc := r.FindStringIndex(s.remainingInput)
	if loc != nil {
		matched := s.remainingInput[:loc[1]]
		s.move(loc[1])
		return &matched
	}

	return nil
}

func (s *Scanner[T]) Skip() {
	if s.skipRegex != nil && s.position < len(s.input) &&
		(s.input[s.position] <= ' ' || s.input[s.position] == '/') {
		s.MatchRegexp(s.skipRegex)
	}
}

// GetPosition returns the current position in the input
func (s *Scanner[T]) GetPosition() int {
	return s.position
}

// GetInput returns the full input string
func (s *Scanner[T]) GetInput() string {
	return s.input
}

// Substring returns the substring from start to end positions
func (s *Scanner[T]) Substring(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(s.input) {
		end = len(s.input)
	}
	if start >= end {
		return ""
	}
	return s.input[start:end]
}
