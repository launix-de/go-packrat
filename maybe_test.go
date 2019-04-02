package packrat

import "testing"

func TestMaybe(t *testing.T) {
	input := "Hello"
	scanner := NewScanner(input, true)

	helloParser := NewAtomParser("Hello", true)
	helloAndWorldParser := NewMaybeParser(helloParser)

	n, err := ParsePartial(helloAndWorldParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != helloAndWorldParser {
			t.Error("Maybe combinator creates node with wrong parser")
		}
		if n.Matched != input {
			t.Error("Maybe combinator doesn't match complete input")
		}
	}

	irregularInput := "Sonne"
	irregularScanner := NewScanner(irregularInput, true)
	irregularParser := NewMaybeParser(helloParser)

	in, ierr := ParsePartial(irregularParser, irregularScanner)
	if ierr != nil {
		t.Error("Maybe combinator doesn't match irregular input")
	}
	if len(in.Children) != 0 {
		t.Error("Maybe combinator doesn't produce zero children for irregular input")
	}
}
