package packrat

import (
	"regexp"
	"strings"
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

	skipRegex *regexp.Regexp
}

func (s *Scanner) Copy() *Scanner {
	return &Scanner{input: s.input, remainingInput: s.remainingInput, position: s.position, memoization: s.memoization, skipRegex: s.skipRegex}
}

var skipWhitespaceRegex = regexp.MustCompile("^[\r\n\t ]+")

func NewScanner(input string, skipWhitespace bool) *Scanner {
	s := &Scanner{input: input, position: 0, memoization: make([]map[Parser]scannerNode, len(input)+1)}
	for i := 0; i < len(s.input)+1; i++ {
		s.memoization[i] = make(map[Parser]scannerNode)
	}
	s.remainingInput = s.input
	if skipWhitespace {
		s.skipRegex = skipWhitespaceRegex
	}
	return s
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

func (s *Scanner) MatchString(str string) *string {
	if strings.HasPrefix(s.remainingInput, str) {
		s.updatePosition(str)
		return &str
	}

	return nil
}

func (s *Scanner) Skip() {
	if s.skipRegex != nil {
		s.MatchRegexp(s.skipRegex)
	}
}
