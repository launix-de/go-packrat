package packrat

import "testing"

func TestJSON(t *testing.T) {
	input := `{"hallo": "3", "welt": true, "subObject": {"hallo": 5}, "subArray": [1,2,3,4,5,6,7],
		"tek": "lel test \" halolaleö pops"}`
	scanner := NewScanner(input, true)

	stringParser := NewAndParser(NewAtomParser(`"`, true), NewRegexParser(`(?:[^"\\]|\\.)*`, false), NewAtomParser(`"`, false))
	valueParser := NewOrParser(nil)
	propParser := NewAndParser(stringParser, NewAtomParser(":", true), valueParser)

	objParser := NewAndParser(NewAtomParser("{", true), NewKleeneParser(propParser, NewAtomParser(",", true)), NewAtomParser("}", true))
	numParser := NewRegexParser(`-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?`, true)
	boolParser := NewRegexParser("(true|false)", true)
	arrayParser := NewAndParser(NewAtomParser("[", true), NewKleeneParser(valueParser, NewAtomParser(",", true)), NewAtomParser("]", true))
	valueParser.Set(objParser, stringParser, numParser, boolParser, arrayParser)

	n, err := Parse(valueParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Matched != input {
			t.Error("JSON combinator doesn't match complete input")
		}
	}
}
