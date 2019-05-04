/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import (
	"unicode"
	"regexp"
)

type scannerNode struct {
	Scanner *Scanner
	Node    Node
}

type Scanner struct {
	input          string
	remainingInput string
	position       int
	memoization    []map[Parser]scannerNode
	breaks []bool

	skipRegex *regexp.Regexp
}

func (s *Scanner) Copy() *Scanner {
	ns := *s
	return &ns
}

var skipWhitespaceRegex = regexp.MustCompile("^[\r\n\t ]+")
var blockbreakRegex = regexp.MustCompile(`\b`)

func NewScanner(input string, skipWhitespace bool) *Scanner {
	s := &Scanner{input: input, position: 0, memoization: make([]map[Parser]scannerNode, len(input)+1)}
	for i := 0; i < len(s.input)+1; i++ {
		s.memoization[i] = make(map[Parser]scannerNode)
	}
	s.remainingInput = s.input
	if skipWhitespace {
		s.skipRegex = skipWhitespaceRegex
	}
	breaks := make([]bool, len(input)+1)
	
	previousWord := false
	for pos, r := range input {
		currentWord := unicode.In(r, unicode.N, unicode.L, unicode.Pc)
		if !currentWord || !previousWord {
			breaks[pos] = true
		}

		previousWord = currentWord
	}
	breaks[len(input)] = true
	s.breaks = breaks

	return s
}

func (s *Scanner) isAtBreak() bool {
	return s.breaks[s.position]
}

func (s *Scanner) updatePosition(reads string) {
	l := len(reads)
	if l > 0 {
		s.remainingInput = s.remainingInput[l:]
		s.position += l
	}
}

func (s *Scanner) MatchRegexp(r *regexp.Regexp) *string {
	matched := r.FindStringSubmatch(s.remainingInput)
	if matched != nil {
		s.updatePosition(matched[0])
		return &matched[0]
	}

	return nil
}

func (s *Scanner) Skip() {
	if s.skipRegex != nil {
		s.MatchRegexp(s.skipRegex)
	}
}
