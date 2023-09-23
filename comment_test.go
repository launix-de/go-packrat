/*
	(c) 2019, 2023 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge
	Author: Carl-Philip Hänsch

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "testing"

func TestComments(t *testing.T) {
	input := "HELLO /* this is a comment */ world"
	scanner := NewScanner(input, SkipWhitespaceAndCommentsRegex)

	helloParser := NewAtomParser("Hello", true, true)
	worldParser := NewAtomParser("World", true, true)
	helloAndWorldParser := NewAndParser(helloParser, worldParser)

	n, err := Parse(helloAndWorldParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Parser != helloAndWorldParser {
			t.Error("And combinator creates node with wrong parser")
		}
	}
}

