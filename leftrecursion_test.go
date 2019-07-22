package packrat

import "testing"

func TestLeftRecursion(t *testing.T) {
	input := "5-1-4-3"
	scanner := NewScanner(input, true)

	emptyParser1 := NewEmptyParser()

	numParser := NewRegexParser(`\d+`, false, true)
	numCombo1 := NewAndParser(emptyParser1, emptyParser1, numParser)
	minusParser := NewAtomParser(`-`, false, true)

	termParser := NewAndParser()
	exprParser := NewOrParser(termParser, numCombo1)
	termParser.Set(exprParser, minusParser, numCombo1)

	n, err := Parse(exprParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != exprParser {
			t.Error("Term parser creates node with wrong parser")
		}
	}
}
