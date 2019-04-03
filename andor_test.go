/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "testing"

func TestAnd(t *testing.T) {
	input := "Hello World"
	scanner := NewScanner(input, true)

	helloParser := NewAtomParser("Hello", true)
	worldParser := NewAtomParser("World", true)
	helloAndWorldParser := NewAndParser(helloParser, worldParser)

	n, err := Parse(helloAndWorldParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != helloAndWorldParser {
			t.Error("And combinator creates node with wrong parser")
		}
		if n.Matched != "HelloWorld" {
			t.Error("And combinator doesn't match complete input")
		}
	}

	irregularInput := "Hello"
	irregularScanner := NewScanner(irregularInput, true)
	irregularParser := NewAndParser(helloParser, worldParser)

	_, ierr := Parse(irregularParser, irregularScanner)
	if ierr == nil {
		t.Error("And combinator matches irregular input")
	}
}

func TestOr(t *testing.T) {
	input := "World"
	scanner := NewScanner(input, true)

	helloParser := NewAtomParser("Hello", true)
	worldParser := NewAtomParser("World", true)
	helloAndWorldParser := NewOrParser(helloParser, worldParser)

	n, err := Parse(helloAndWorldParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != helloAndWorldParser {
			t.Error("Or combinator creates node with wrong parser")
		}
		if n.Matched != input {
			t.Error("Or combinator doesn't match complete input")
		}
	}

	irregularInput := "Sonne"
	irregularScanner := NewScanner(irregularInput, true)
	irregularParser := NewAndParser(helloParser, worldParser)

	_, ierr := Parse(irregularParser, irregularScanner)
	if ierr == nil {
		t.Error("Or combinator matches irregular input")
	}
}
