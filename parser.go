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
	Match(s *Scanner) (*Scanner, Node)
}

var emptyString = ""

type Node struct {
	Matched  string
	Parser   Parser
	Children []Node
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

func match(os *Scanner, p Parser) (*Scanner, Node){
	startPosition := os.position 
	if os.position >= len(os.input) {
		return nil, Node{} 
	}
	
	m, mExists := os.memoization[startPosition]
	if !mExists {
		m = make(map[Parser]scannerNode)
		os.memoization[startPosition] = m
	}

	cached, wasCached := m[p]
	if wasCached {
		if cached.Scanner == nil {
			m[p] = cached
		}
		nss, node := cached.Scanner, cached.Node
		return nss, node
	}

	s := os.Copy()
	ns, n := p.Match(s)
	m[p] = scannerNode{Scanner: ns, Node: n}

	return ns, n
}

func ParsePartial(p Parser, originalScanner *Scanner) (*Node, *ParserError) {
	newScanner, node := match(originalScanner, p)
	if newScanner != nil {
		return &node, nil
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

func Parse(p Parser, originalScanner *Scanner) (*Node, *ParserError) {
	newScanner, node := match(originalScanner, p)
	if newScanner != nil {
		if len(newScanner.remainingInput) > 0 {
			consumed := originalScanner.input[:newScanner.position]
			line := strings.Count(consumed, "\n") + 1
			column := newScanner.position - strings.LastIndex(consumed, "\n") + 1
			e := &ParserError{Parser: p, Line: line, Column: column, Position: newScanner.position, Input: originalScanner.input}
			return nil, e
		}

		return &node, nil
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
