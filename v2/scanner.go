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
	nextMemo uint32 // index into memoSlabs (1-based, 0 = end of chain)
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

// positions stores per-position parsing state as packed uint32:
//   bits  0-23: memo entry index (1-based into memoSlabs, 0 = none)
//   bits 24-31: head entry index (1-based into headPtrs, 0 = none)
// 32 bits per position = 16 positions per cache line.

type Scanner[T any] struct {
	input           string
	remainingInput  string
	position        int
	positions       []uint32
	invocationStack *Lr[T]
	breaks          []bool

	headpool        sync.Pool
	lrPool          sync.Pool
	headPtrs        []*Head[T] // small array for head pointer lookup (indexed by head bits - 1)

	// MemoEntry slab allocator: multiple slabs of 256 entries each.
	// Slabs are never moved, so pointers into them remain valid.
	memoSlabs       [][]MemoEntry[T]
	memoCount       uint32 // total entries allocated across all slabs

	skipRegex *regexp.Regexp
}

const memoSlabSize = 256

// position slot helpers
func posMemo(slot uint32) uint32 { return slot & 0x00FFFFFF }
func posHead(slot uint32) uint8  { return uint8(slot >> 24) }

func (s *Scanner[T]) newMemoEntry() (uint32, *MemoEntry[T]) {
	idx := s.memoCount
	slabIdx := idx >> 8 // idx / 256
	if int(slabIdx) >= len(s.memoSlabs) {
		s.memoSlabs = append(s.memoSlabs, make([]MemoEntry[T], memoSlabSize))
	}
	s.memoCount++
	oneBasedIdx := idx + 1
	m := &s.memoSlabs[slabIdx][idx&0xFF]
	return oneBasedIdx, m
}

func (s *Scanner[T]) memoAt(oneBasedIdx uint32) *MemoEntry[T] {
	idx := oneBasedIdx - 1
	return &s.memoSlabs[idx>>8][idx&0xFF]
}

// Copy clones the scanner state. Slices are shared (read-only after construction).
// Pools are shared by copying the New functions.
func (s *Scanner[T]) Copy() *Scanner[T] {
	ns := &Scanner[T]{
		input:           s.input,
		remainingInput:  s.remainingInput,
		position:        s.position,
		positions:       s.positions,
		invocationStack: s.invocationStack,
		breaks:          s.breaks,
		headPtrs:        s.headPtrs,
		memoSlabs:       s.memoSlabs,
		memoCount:       s.memoCount,
		skipRegex:       s.skipRegex,
	}
	ns.headpool.New = s.headpool.New
	ns.lrPool.New = s.lrPool.New
	return ns
}

func (s *Scanner[T]) memoLookup(pos int, rule Parser[T]) *MemoEntry[T] {
	for idx := posMemo(s.positions[pos]); idx != 0; {
		m := s.memoAt(idx)
		if m.rule == rule {
			return m
		}
		idx = m.nextMemo
	}
	return nil
}

func (s *Scanner[T]) Recall(rule Parser[T], pos int) *MemoEntry[T] {
	m := s.memoLookup(pos, rule)

	// If not growing a seed parse, just return what is stored in the memo table
	headIdx := posHead(s.positions[pos])
	if headIdx == 0 {
		return m
	}
	head := s.headPtrs[headIdx-1]

	// Do not evaluate any rule that is not involved in this left recursion
	if m == nil && !head.IsInvolved(rule) {
		_, me := s.newMemoEntry()
		me.Position = s.position
		return me
	}

	// Allow involved rules to be evaluated, but only once, during a seed-growing iteration
	if head.IsEvaluated(rule) {
		delete(head.evalSet, rule)
		node, ok := rule.Match(s)
		_, me := s.newMemoEntry()
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

func (s *Scanner[T]) setHead(pos int, h *Head[T]) {
	s.headPtrs = append(s.headPtrs, h)
	headIdx := uint32(len(s.headPtrs))
	s.positions[pos] = posMemo(s.positions[pos]) | (headIdx << 24)
}

func (s *Scanner[T]) clearHead(pos int) {
	s.positions[pos] = posMemo(s.positions[pos])
}

func (s *Scanner[T]) GrowLr(rule Parser[T], p int, m *MemoEntry[T], h *Head[T]) (Node[T], bool) {
	s.setHead(p, h)
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
	s.headpool.Put(h)
	s.clearHead(p)
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
	s := &Scanner[T]{input: input, position: 0, positions: make([]uint32, len(input)+1)}
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

// Reset reinitializes the scanner for a new input, reusing allocated slices
// when possible. This allows pooling Scanners across queries to avoid per-query
// construction allocations.
func (s *Scanner[T]) Reset(input string, skipper *regexp.Regexp) {
	s.input = input
	s.position = 0
	s.remainingInput = input
	s.skipRegex = skipper
	s.invocationStack = nil
	s.memoCount = 0
	s.headPtrs = s.headPtrs[:0]

	needed := len(input) + 1

	// Reuse positions slice if large enough
	if cap(s.positions) >= needed {
		s.positions = s.positions[:needed]
		for i := range s.positions {
			s.positions[i] = 0
		}
	} else {
		s.positions = make([]uint32, needed)
	}

	// Reuse breaks slice if large enough
	if cap(s.breaks) >= needed {
		s.breaks = s.breaks[:needed]
		for i := range s.breaks {
			s.breaks[i] = false
		}
	} else {
		s.breaks = make([]bool, needed)
	}

	// Rebuild word breaks
	previousWord := false
	for pos, r := range input {
		currentWord := unicode.In(r, unicode.N, unicode.L, unicode.Pc)
		if !currentWord || !previousWord {
			s.breaks[pos] = true
		}
		previousWord = currentWord
	}
	s.breaks[len(input)] = true
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
