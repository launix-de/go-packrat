/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "testing"

func TestJSON(t *testing.T) {
	input := `{"menu": {
		"header": "SVG Viewer",
		"items": [
			{"id": "Open"},
			{"id": "OpenNew", "label": "Open New"},
			null,
			{"id": "ZoomIn", "label": "Zoom In"},
			{"id": "ZoomOut", "label": "Zoom Out"},
			{"id": "OriginalView", "label": "Original View"},
			null,
			{"id": "Quality"},
			{"id": "Pause"},
			{"id": "Mute"},
			null,
			{"id": "Find", "label": "Find..."},
			{"id": "FindAgain", "label": "Find Again"},
			{"id": "Copy"},
			{"id": "CopyAgain", "label": "Copy Again"},
			{"id": "CopySVG", "label": "Copy SVG"},
			{"id": "ViewSVG", "label": "View SVG"},
			{"id": "ViewSource", "label": "View Source"},
			{"id": "SaveAs", "label": "Save As"},
			null,
			{"id": "Help"},
			{"id": "About", "label": "About Adobe CVG Viewer..."}
		]
	}}`
	scanner := NewScanner(input, true)

	stringParser := NewAndParser(NewAtomParser(`"`, false, true), NewRegexParser(`(?:[^"\\]|\\.)*`, false, false), NewAtomParser(`"`, false, false))
	valueParser := NewOrParser(nil)
	propParser := NewAndParser(stringParser, NewAtomParser(":", false, true), valueParser)

	objParser := NewAndParser(NewAtomParser("{", false, true), NewKleeneParser(propParser, NewAtomParser(",", false, true)), NewAtomParser("}", false, true))
	nullParser := NewAtomParser("null", false, true)
	numParser := NewRegexParser(`-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?`, false, true)
	boolParser := NewRegexParser("(true|false)", false, true)
	arrayParser := NewAndParser(NewAtomParser("[", false, true), NewKleeneParser(valueParser, NewAtomParser(",", false, true)), NewAtomParser("]", false, true))
	valueParser.Set(nullParser, objParser, stringParser, numParser, boolParser, arrayParser)

	_, err := Parse(valueParser, scanner)
	if err != nil {
		t.Error(err)
	}
}
