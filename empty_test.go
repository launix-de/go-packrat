package packrat

import "testing"

func TestEmptyParser(t *testing.T) {
	input := "Hello"
	scanner := NewScanner(input, true)

	emptyParser := NewEmptyParser()
	helloParser := NewAtomParser("Hello", false, true)
	helloAndWorldParser := NewAndParser(emptyParser, helloParser)

	n, err := Parse(helloAndWorldParser, scanner)
	if err != nil {
		t.Error(err)
	}

	if n.Parser != helloAndWorldParser {
		t.Error("Empty Test combinator creates node with wrong parser")
	}
	if n.Matched != input {
		t.Error("Empty Test combinator doesn't match complete input")
	}
	if len(n.Children) != 2 {
		t.Error("Empty Test combinator doesn't produce 3 children")
	}
	if n.Children[0].Parser != emptyParser || n.Children[1].Parser != helloParser {
		t.Error("Empty Test combinator AST nodes do not point to their respective parsers")
	}
	if n.Children[0].Matched != "" || n.Children[1].Matched != "Hello" {
		t.Error("Empty Test combinator sub parsers match wrong input: '" + n.Children[0].Matched + "' '" + n.Children[1].Matched + "'")
	}
}
