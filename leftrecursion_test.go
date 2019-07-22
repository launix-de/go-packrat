package packrat

import "testing"

func TestLeftRecursion(t *testing.T) {
	input := "5-1-4-3"
	scanner := NewScanner(input, true)

	numParser := NewRegexParser(`\d+`, false, true)
	minusParser := NewAtomParser(`-`, false, true)

	termParser := NewAndParser()
	exprParser := NewOrParser(termParser, numParser)
	termParser.Set(exprParser, minusParser, numParser)

	n, err := Parse(exprParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != exprParser {
			t.Error("Term parser creates node with wrong parser")
		}
	}
}
