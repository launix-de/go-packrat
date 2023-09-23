/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "testing"

func TestManySeparator(t *testing.T) {
	input := "Hello, Hello"
	scanner := NewScanner(input, SkipWhitespaceRegex)

	helloParser := NewAtomParser("Hello", false, true)
	helloAndWorldParser := NewManyParser(helloParser, NewAtomParser(",", false, true))

	n, err := Parse(helloAndWorldParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != helloAndWorldParser {
			t.Error("Many combinator creates node with wrong parser")
		}
		if n.Matched != input {
			t.Error("Many combinator doesn't match complete input")
		}
		if len(n.Children) != 3 {
			t.Error("Many combinator doesn't produce 3 children")
		}
		if n.Children[0].Matched != "Hello" || n.Children[1].Matched != "," || n.Children[2].Matched != "Hello" {
			t.Error("Many combinator sub parsers match wrong input: '" + n.Children[0].Matched + "' '" + n.Children[1].Matched + "' '" + n.Children[2].Matched + "'")
		}
	}

	irregularScanner := NewScanner("Hello, Hello, Hello, ", SkipWhitespaceRegex)
	irregularParser := NewManyParser(helloParser, NewAtomParser(",", false, true))

	_, ierr := Parse(irregularParser, irregularScanner)
	if ierr == nil {
		t.Error("Many combinator matches irregular input")
	}
}

func TestManySeparatorRegex(t *testing.T) {
	input := "   23, 45"
	scanner := NewScanner(input, SkipWhitespaceRegex)

	helloParser := NewRegexParser(`\d+`, false, true)
	helloAndWorldParser := NewManyParser(helloParser, NewAtomParser(",", false, true))

	n, err := Parse(helloAndWorldParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != helloAndWorldParser {
			t.Error("Many combinator creates node with wrong parser")
		}
		if n.Matched != "23, 45" {
			t.Error("Many combinator doesn't match complete input")
		}
		if len(n.Children) != 3 {
			t.Error("Many combinator doesn't produce 3 children")
		}
		if n.Children[0].Matched != "23" || n.Children[1].Matched != "," || n.Children[2].Matched != "45" {
			t.Error("Many combinator sub parsers match wrong input: '" + n.Children[0].Matched + "' '" + n.Children[1].Matched + "' '" + n.Children[2].Matched + "'")
		}
	}

	irregularScanner := NewScanner("Hello, Hello, Hello, ", SkipWhitespaceRegex)
	irregularParser := NewManyParser(helloParser, NewAtomParser(",", false, true))

	_, ierr := Parse(irregularParser, irregularScanner)
	if ierr == nil {
		t.Error("Many combinator matches irregular input")
	}
}

func TestMany(t *testing.T) {
	input := "Hello Hello Hello"
	scanner := NewScanner(input, SkipWhitespaceRegex)

	helloParser := NewAtomParser("Hello", false, true)
	helloAndWorldParser := NewManyParser(helloParser, nil)

	n, err := Parse(helloAndWorldParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != helloAndWorldParser {
			t.Error("Many combinator creates node with wrong parser")
		}
		if len(n.Children) != 3 {
			t.Error("Many combinator doesn't produce 3 children")
		}
	}

	irregularScanner := NewScanner("Sonne", SkipWhitespaceRegex)
	irregularParser := NewManyParser(helloParser, nil)

	_, ierr := ParsePartial(irregularParser, irregularScanner)
	if ierr == nil {
		t.Error("Many combinator matches irregular input")
	}

	irregularScanner2 := NewScanner("", SkipWhitespaceRegex)
	irregularParser2 := NewManyParser(helloParser, nil)

	_, ierr = Parse(irregularParser2, irregularScanner2)
	if ierr == nil {
		t.Error("Many combinator matches irregular input")
	}
}
