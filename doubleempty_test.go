package packrat

import "testing"

func TestDoubleEmpty(t *testing.T) {
	input := ""
	scanner := NewScanner(input, true)

	emptyParser := NewEmptyParser()
	termParser := NewAndParser(emptyParser, emptyParser)

	n, err := Parse(termParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != termParser {
			t.Error("Term parser creates node with wrong parser")
		}
	}
}
