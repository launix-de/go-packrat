package packrat

import "testing"

func TestEndParser(t *testing.T) {
	input := "Hello"
	scanner := NewScanner(input, true)

	endParser := NewEndParser(false)
	helloParser := NewAtomParser("Hello", false, true)
	helloAndWorldParser := NewAndParser(helloParser, endParser)

	n, err := Parse(helloAndWorldParser, scanner)
	if err != nil {
		t.Error(err)
	}

	if n.Parser != helloAndWorldParser {
		t.Error("End Test combinator creates node with wrong parser")
	}
	if n.Matched != input {
		t.Error("End Test combinator doesn't match complete input")
	}
	if len(n.Children) != 2 {
		t.Error("End Test combinator doesn't produce 3 children")
	}
	if n.Children[0].Parser != helloParser || n.Children[1].Parser != endParser {
		t.Error("End Test combinator AST nodes do not point to their respective parsers")
	}
	if n.Children[0].Matched != "Hello" || n.Children[1].Matched != "" {
		t.Error("End Test combinator sub parsers match wrong input: '" + n.Children[0].Matched + "' '" + n.Children[1].Matched + "'")
	}
}
