/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import (
	"fmt"
	"strings"
)

type Parser interface {
	Match(s *Scanner) *Node
	Description(visited map[Parser]bool) string
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
	startpos := e.Position - 1
	if startpos < 0 {
		startpos = 0
	}

	endpos := e.Position + 10
	if endpos >= len(e.Input) {
		endpos = len(e.Input) - 1
	}
	s := e.Input[startpos:endpos]
	return fmt.Sprintf("Parser failed at line %d, column %d (position %d of input string) near %s", e.Line, e.Column, e.Position, s)
}

func ParsePartial(p Parser, originalScanner *Scanner) (*Node, *ParserError) {
	node := originalScanner.applyRule(p)
	if node != nil {
		return node, nil
	}

	maxPos := 0
	var failedParsers []Parser
	for index := len(originalScanner.input) - 1; index > 0; index-- {
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
	column := maxPos - strings.LastIndex(consumed, "\n") + 1
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
	for index := len(originalScanner.memoization) - 1; index > 0; index-- {
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
	column := maxPos - strings.LastIndex(consumed, "\n") + 1
	e := &ParserError{FailedParsers: failedParsers, Parser: p, Line: line, Column: column, Position: maxPos, Input: originalScanner.input}

	return nil, e
}

func writeDebug(p Parser, stack map[Parser]bool, subs ...Parser) string {
	stack[p] = true
	b := strings.Builder{}

	for i, s := range subs {
		if i > 0 {
			b.WriteString(", ")
		}
		subVisited := stack[s]
		if subVisited {
			b.WriteString("--- parent ---")
		} else {
			b.WriteString(s.Description(stack))
		}
	}

	stack[p] = false
	return b.String()
}
