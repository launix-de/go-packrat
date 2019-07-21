/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import (
	"regexp"
	"unicode"
)

type MemoEntry struct {
	Lr  *Lr
	Ans *Node

	Position int
}

type Head struct {
	rule        Parser
	involvedSet map[Parser]bool
	evalSet     map[Parser]bool
}

func NewHead(rule Parser) *Head {
	return &Head{rule: rule, involvedSet: make(map[Parser]bool), evalSet: make(map[Parser]bool)}
}

func (h *Head) IsInvolved(rule Parser) bool {
	if rule == h.rule {
		return true
	}

	_, isInvoled := h.involvedSet[rule]
	return isInvoled
}

func (h *Head) IsEvaluated(rule Parser) bool {
	_, isEvaluated := h.evalSet[rule]
	return isEvaluated
}

type Lr struct {
	seed *Node
	rule Parser
	head *Head
	next *Lr
}

type Scanner struct {
	input           string
	remainingInput  string
	position        int
	memoization     map[int]map[Parser]*MemoEntry
	heads           map[int]*Head
	invocationStack *Lr
	breaks          map[int]bool

	skipRegex *regexp.Regexp
}

// Copy clones the scanner state. Memoization and break maps are shared
func (s *Scanner) Copy() *Scanner {
	ns := *s
	return &ns
}

func (s *Scanner) Recall(rule Parser, pos int) *MemoEntry {
	mmap, mmapExists := s.memoization[pos]
	var m *MemoEntry
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
		return nil
	}

	// Allow involved rules to be evaluated, but only once, during a seed-growing iteration
	if head.IsEvaluated(rule) {
		delete(head.evalSet, rule)
		node := rule.Match(s)
		return &MemoEntry{Position: s.position, Ans: node}
	}

	return m
}

func (s *Scanner) SetupLr(rule Parser, l *Lr) {
	if l.head == nil {
		l.head = NewHead(rule)
	}
	stack := s.invocationStack
	for stack.head != l.head {
		stack.head = l.head
		l.head.involvedSet[stack.rule] = true
		stack = stack.next
	}
}

func (s *Scanner) GrowLr(rule Parser, p int, m *MemoEntry, h *Head) *Node {
	s.heads[p] = h
	for {
		s.setPosition(p)
		h.evalSet = make(map[Parser]bool)
		for k, v := range h.involvedSet {
			h.evalSet[k] = v
		}
		ans := rule.Match(s)
		if ans == nil || s.position <= m.Position {
			break
		}
		m.Ans = ans
		m.Position = s.position
	}
	delete(s.heads, p)
	s.setPosition(m.Position)
	return m.Ans
}

func (s *Scanner) LrAnswer(rule Parser, pos int, m *MemoEntry) *Node {
	h := m.Lr.head
	if h.rule != rule {
		return m.Lr.seed
	}
	m.Ans = m.Lr.seed
	m.Lr = nil

	if m.Ans == nil {
		return nil
	}
	return s.GrowLr(rule, pos, m, h)
}

var skipWhitespaceRegex = regexp.MustCompile("^[\r\n\t ]+")

func NewScanner(input string, skipWhitespace bool) *Scanner {
	s := &Scanner{input: input, position: 0, memoization: make(map[int]map[Parser]*MemoEntry),
		heads: make(map[int]*Head)}
	s.remainingInput = s.input
	if skipWhitespace {
		s.skipRegex = skipWhitespaceRegex
	}
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

func (s *Scanner) isAtBreak() bool {
	return s.breaks[s.position]
}

func (s *Scanner) move(n int) {
	s.position += n
	s.remainingInput = s.input[s.position:]
}

func (s *Scanner) setPosition(n int) {
	s.position = n
	s.remainingInput = s.input[s.position:]
}

func (s *Scanner) MatchRegexp(r *regexp.Regexp) *string {
	matched := r.FindStringSubmatch(s.remainingInput)
	if matched != nil {
		s.move(len(matched[0]))
		return &matched[0]
	}

	return nil
}

func (s *Scanner) Skip() {
	if s.skipRegex != nil {
		s.MatchRegexp(s.skipRegex)
	}
}
