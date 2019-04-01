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

	stringParser := NewAndParser(NewAtomParser(`"`, true), NewRegexParser(`(?:[^"\\]|\\.)*`, false), NewAtomParser(`"`, false))
	valueParser := NewOrParser(nil)
	propParser := NewAndParser(stringParser, NewAtomParser(":", true), valueParser)

	objParser := NewAndParser(NewAtomParser("{", true), NewKleeneParser(propParser, NewAtomParser(",", true)), NewAtomParser("}", true))
	nullParser := NewAtomParser("null", true)
	numParser := NewRegexParser(`-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?`, true)
	boolParser := NewRegexParser("(true|false)", true)
	arrayParser := NewAndParser(NewAtomParser("[", true), NewKleeneParser(valueParser, NewAtomParser(",", true)), NewAtomParser("]", true))
	valueParser.Set(nullParser, objParser, stringParser, numParser, boolParser, arrayParser)

	n, err := Parse(valueParser, scanner)
	if err != nil {
		t.Error(err)
	} else {
		if n.Matched != input {
			t.Error("JSON combinator doesn't match complete input")
		}
	}
}
