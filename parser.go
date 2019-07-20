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
)

type Parser interface {
	Match(s *Scanner) (*Scanner, Node)
	Description(visited map[Parser]bool) string
}

func match(os *Scanner, p Parser) (*Scanner, Node) {
	startPosition := os.position

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

	var triedParsers map[Parser]bool
	triedParsers = os.triedParsers[startPosition]
	if triedParsers == nil {
		triedParsers = make(map[Parser]bool)
		os.triedParsers[startPosition] = triedParsers
	}
	_, alreadyTried := triedParsers[p]
	if alreadyTried {
		if os.lrDetected.Parser != nil && os.lrDetected.Parser != p {
			panic("Indirect left recursion is not supported")
		}
		os.lrDetected.Parser = p
		os.lrDetected.StartPos = os.position
		visited := make(map[Parser]bool)
		fmt.Println("Left recursion detected at pos " + strconv.Itoa(os.position) + ": " + os.lrDetected.Parser.Description(visited))

		// Ablehnen, damit der Samen angelegt werden kann
		return nil, Node{}
	}
	triedParsers[p] = true

	s := os.Copy()
	ns, n := p.Match(s)

	/**
	Term -> Term "-" Term / Num
	Num -> \d+

	5 - 4 - 3
	*/

	m[p] = scannerNode{Scanner: ns, Node: n}

	if os.lrDetected.Parser != nil && os.lrDetected.Parser == p && !os.lrDetected.InDescent {
		// lp := ns.lrDetected
		fmt.Println("Outer left recursion: " + n.Matched)

		// Jetzt den Samen wachsen lassen - den Parser immer wieder aufrufen, bis er nicht mehr voran kommt
		//		changedPosition := false

		nns := ns.Copy()
		nns.lrDetected.InDescent = false
		for {
			nns.move(-(nns.position - s.lrDetected.StartPos))
			nns2, n2 := match(nns, p)

			if nns2 == nil {
				return nns, n
			}
			nns, n = nns2, n2
		}
	}
	return ns, n
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

func ParsePartial(p Parser, originalScanner *Scanner) (*Node, *ParserError) {
	newScanner, node := match(originalScanner, p)
	if newScanner != nil {
		return &node, nil
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
