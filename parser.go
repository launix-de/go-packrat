package packrat

import (
	"errors"
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

	return nil, errors.New("Parser did not match")
}

func Parse(p Parser, originalScanner *Scanner) (*Node, error) {
	newScanner, node := p.Match(originalScanner)
	if newScanner != nil {
		if len(newScanner.remainingInput) > 0 {
			return nil, errors.New("Parser did not match complete input")
		}

		return &node, nil
	}

	return nil, errors.New("Parser did not match")
}
