package packrat

import "testing"

func TestLeftRecursion(t *testing.T) {
	input := "5 - 3 - 1"
	scanner := NewScanner(input, true)

	numParser := NewRegexParser(`\d+`, false, true)
	minusParser := NewAtomParser(`-`, false, true)

	termParser := NewAndParser()
	exprParser := NewOrParser(termParser, numParser)
	termParser.Set(termParser, minusParser, numParser)

	n, err := Parse(exprParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != termParser {
			t.Error("Term parser creates node with wrong parser")
		}
	}
}