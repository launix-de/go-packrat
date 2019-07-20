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

type scannerNode struct {
	Scanner *Scanner
	Node    Node
}

type Scanner struct {
	input           string
	remainingInput  string
	position        int
	memoization     map[int]map[Parser]scannerNode
	invocationStack []Parser
	breaks          map[int]bool

	skipRegex *regexp.Regexp
}

// Copy clones the scanner state. Memoization and break maps are shared
func (s *Scanner) Copy() *Scanner {
	ns := *s

	ns.invocationStack = make([]Parser, len(s.invocationStack))
	copy(ns.invocationStack, s.invocationStack)

	return &ns
}

var skipWhitespaceRegex = regexp.MustCompile("^[\r\n\t ]+")

func NewScanner(input string, skipWhitespace bool) *Scanner {
	s := &Scanner{input: input, position: 0, memoization: make(map[int]map[Parser]scannerNode)}
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
