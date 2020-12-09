/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Parser interface {
	Match(s *Scanner) *Node
}

func (s *Scanner) applyRule(rule Parser) *Node {
	startPosition := s.position

	memmap, memmapExists := s.memoization[startPosition]
	if !memmapExists {
		memmap = make(map[Parser]*MemoEntry)
		s.memoization[startPosition] = memmap
	}

	m := s.Recall(rule, startPosition)
	if m == nil {
		lr := &Lr{seed: nil, rule: rule, head: nil, next: s.invocationStack}
		s.invocationStack = lr
		m := &MemoEntry{Lr: lr, Position: startPosition}
		memmap[rule] = m
		ans := rule.Match(s)
		s.invocationStack = s.invocationStack.next
		m.Position = s.position
		if lr.head != nil {
			lr.seed = ans
			return s.LrAnswer(rule, startPosition, m)
		}

		m.Lr = nil
		m.Ans = ans
		return ans
	}

	s.setPosition(m.Position)

	if m.Lr != nil {
		s.SetupLr(rule, m.Lr)
		return m.Lr.seed
	}

	return m.Ans
}

var emptyString = ""

type Node struct {
	Matched  string
	Parser   Parser
	Children []*Node
}

type ParserError struct {
	Parser        Parser
	Line          int
	Column        int
	Position      int
	FailedParsers []Parser
	Input         string
}

func (e *ParserError) Error() string {
	linestartpos := e.Position - 30

	startpos := linestartpos
	if startpos < 0 {
		startpos = 0
	}

	endpos := e.Position + 10
	if endpos >= len(e.Input) {
		endpos = len(e.Input)
		if endpos < 0 {
			endpos = 0
		}
	}

	atomParsers := make(map[*AtomParser]bool)
	regexParsers := make(map[*RegexParser]bool)
	eofParser := false
	allskipws := true

	for _, p := range e.FailedParsers {
		switch pa := p.(type) {
		case *AtomParser:
			atomParsers[pa] = true
			if !pa.skipWs {
				allskipws = false
			}
		case *RegexParser:
			regexParsers[pa] = true
			if !pa.skipWs {
				allskipws = false
			}
		case *EndParser:
			eofParser = true
		}
	}

	expected := strings.Builder{}
	count := 0
	for r := range atomParsers {
		if count >= 5 {
			break
		}
		expected.WriteString("- " + r.atom)
		if r.skipWs {
			expected.WriteString(" (with leading whitespace)")
		}
		expected.WriteString("\r\n")
		count++
	}
	for r := range regexParsers {
		if count >= 5 {
			break
		}
		expected.WriteString("- Regex: " + r.rs)
		if r.skipWs {
			expected.WriteString(" (with leading whitespace)")
		}
		expected.WriteString("\r\n")
		count++
	}
	if eofParser && count < 5 {
		expected.WriteString("- End of input\r\n")
	}

	epos := e.Position
	for allskipws && epos < len(e.Input)-1 && unicode.IsSpace(rune(e.Input[epos])) {
		epos++
	}
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Parser failed at line %d, column %d (position %d of input string).", e.Line, e.Column, e.Position+1))
	replacer := strings.NewReplacer("\r\n", "\\n", "\n", "\\n", "\t", "  ")
	builder.WriteString("\r\n" + replacer.Replace(e.Input[startpos:endpos]))
	builder.WriteString("\r\n")
	runes := []rune(e.Input)
	for i := startpos; i < epos && i < len(runes); i++ {
		c := runes[i]
		switch c {
		case '\n':
			builder.WriteString("  ")
		case '\t':
			builder.WriteString("  ")
		default:
			builder.WriteRune(' ')
		}
	}
	builder.WriteString("^\r\n")
	builder.WriteString("Expected one of " + strconv.Itoa(len(atomParsers)+len(regexParsers)) + " alternatives:\r\n" + expected.String() + "Found: " + strings.ReplaceAll(e.Input[e.Position:endpos], "\n", "\\n"))

	return builder.String()
}

func ParsePartial(p Parser, originalScanner *Scanner) (*Node, *ParserError) {
	node := originalScanner.applyRule(p)
	if node != nil {
		return node, nil
	}

	maxPos := 0
	var failedParsers []Parser
	for index := len(originalScanner.input) - 1; index >= 0; index-- {
		m, mExists := originalScanner.memoization[index]
		if mExists && len(m) > 0 {
			maxPos = index
			for k := range m {
				failedParsers = append(failedParsers, k)
			}
			break
		}
	}

	consumed := originalScanner.input[:maxPos]
	line := strings.Count(consumed, "\n") + 1
	lastBreak := strings.LastIndex(consumed, "\n")
	if lastBreak < 0 {
		lastBreak = 0
	}
	column := maxPos - lastBreak + 1
	e := &ParserError{FailedParsers: failedParsers, Parser: p, Line: line, Column: column, Position: maxPos, Input: originalScanner.input}
	return nil, e
}

func Parse(p Parser, originalScanner *Scanner) (*Node, *ParserError) {
	node := originalScanner.applyRule(p)
	if node != nil {
		if len(originalScanner.remainingInput) > 0 {
			consumed := originalScanner.input[:originalScanner.position]
			line := strings.Count(consumed, "\n") + 1
			column := originalScanner.position - strings.LastIndex(consumed, "\n") + 1
			e := &ParserError{Parser: p, Line: line, Column: column, Position: originalScanner.position, Input: originalScanner.input}
			return nil, e
		}

		return node, nil
	}

	maxPos := 0
	var failedParsers []Parser
	for index := len(originalScanner.input) - 1; index >= 0; index-- {
		m := originalScanner.memoization[index]
		if len(m) > 0 {
			maxPos = index
			for k := range m {
				failedParsers = append(failedParsers, k)
			}
			break
		}
	}

	consumed := originalScanner.input[:maxPos]
	line := strings.Count(consumed, "\n") + 1
	lastBreak := strings.LastIndex(consumed, "\n")
	if lastBreak < 0 {
		lastBreak = 0
	}
	column := maxPos - lastBreak + 1
	e := &ParserError{FailedParsers: failedParsers, Parser: p, Line: line, Column: column, Position: maxPos, Input: originalScanner.input}

	return nil, e
}
