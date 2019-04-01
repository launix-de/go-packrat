package packrat

import (
	"errors"
	"strconv"
)

var emptyString = ""

type Node struct {
	Matched  string
	Parser   Parser
	Children []Node
}

func ParsePartial(p Parser, originalScanner *Scanner) (*Node, error) {
	newScanner, node := p.Match(originalScanner)
	if newScanner != nil {
		return &node, nil
	}

	maxPos := -1
	for index := len(originalScanner.memoization) - 1; index > 0; index-- {
		m := originalScanner.memoization[index]
		if len(m) > 0 {
			maxPos = index
			break
		}
	}

	return nil, errors.New("Parser failed at position " + strconv.Itoa(maxPos))
}

func Parse(p Parser, originalScanner *Scanner) (*Node, error) {
	newScanner, node := p.Match(originalScanner)
	if newScanner != nil {
		if len(newScanner.remainingInput) > 0 {
			return nil, errors.New("Parser did not match complete input")
		}

		return &node, nil
	}

	maxPos := -1
	for index := len(originalScanner.memoization) - 1; index > 0; index-- {
		m := originalScanner.memoization[index]
		if len(m) > 0 {
			maxPos = index
			break
		}
	}

	return nil, errors.New("Parser failed at position " + strconv.Itoa(maxPos))
}
