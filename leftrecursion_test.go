package packrat

import "testing"

func TestLeftRecursion(t *testing.T) {
	input := "5 - 1 - 4"
	scanner := NewScanner(input, true)

	numParser := NewRegexParser(`\d+`, false, true)
	minusParser := NewAtomParser(`-`, false, true)

	termParser := NewAndParser()
	exprParser := NewOrParser(termParser, numParser)
	termParser.Set(exprParser, minusParser, numParser)

	n, err := Parse(termParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != termParser {
			t.Error("Term parser creates node with wrong parser")
		}
	}
}
