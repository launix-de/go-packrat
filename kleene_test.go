/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "testing"

func TestKleene(t *testing.T) {
	input := "Hello Hello Hello"
	scanner := NewScanner(input, true)

	helloParser := NewAtomParser("Hello", true)
	helloAndWorldParser := NewKleeneParser(helloParser, nil)

	n, err := Parse(helloAndWorldParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != helloAndWorldParser {
			t.Error("Kleene combinator creates node with wrong parser")
		}
		if len(n.Children) != 3 {
			t.Error("Kleene combinator doesn't produce 3 children")
		}
	}

	irregularInput := "Sonne"
	irregularScanner := NewScanner(irregularInput, true)
	irregularParser := NewKleeneParser(helloParser, nil)

	in, ierr := ParsePartial(irregularParser, irregularScanner)
	if ierr != nil {
		t.Error("Kleene combinator doesn't match irregular input")
	}
	if len(in.Children) != 0 {
		t.Error("Kleene combinator doesn't produce zero children for irregular input")
	}
}
func TestKleeneSeparator(t *testing.T) {
	input := "Hello, Hello, Hello"
	scanner := NewScanner(input, true)

	helloParser := NewAtomParser("Hello", true)
	sepParser := NewAtomParser(",", true)
	helloAndWorldParser := NewKleeneParser(helloParser, sepParser)

	n, err := Parse(helloAndWorldParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != helloAndWorldParser {
			t.Error("Kleene combinator creates node with wrong parser")
		}
		if n.Matched != "Hello,Hello,Hello" {
			t.Error("Kleene combinator doesn't match complete input")
		}
		if len(n.Children) != 5 {
			t.Error("Kleene combinator doesn't produce 3 children")
		}
	}

	irregularInput := "Sonne"
	irregularScanner := NewScanner(irregularInput, true)
	irregularParser := NewKleeneParser(helloParser, nil)

	in, ierr := ParsePartial(irregularParser, irregularScanner)
	if ierr != nil {
		t.Error("Kleene combinator doesn't match irregular input")
	}
	if len(in.Children) != 0 {
		t.Error("Kleene combinator doesn't produce zero children for irregular input")
	}
}
