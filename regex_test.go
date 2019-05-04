/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "testing"

func TestRegex(t *testing.T) {
	input := "-3.4"
	scanner := NewScanner(input, true)

	numParser := NewRegexParser("-?\\d+\\.\\d+", false, false)

	_, err := Parse(numParser, scanner)
	if err != nil {
		t.Error(err)
	}

	irregularInput := "3,4"
	irregularScanner := NewScanner(irregularInput, true)

	_, ierr := Parse(numParser, irregularScanner)
	if ierr == nil {
		t.Error("Regex combinator matches irregular input")
	}
}
