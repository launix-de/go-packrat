package packrat

import "testing"

func TestOrParserCharMap(t *testing.T) {
	identity := func(s string) string { return s }

	aParser := NewAtomParser[string]("alpha", "alpha", false, false)
	bParser := NewAtomParser[string]("beta", "beta", false, false)
	gParser := NewAtomParser[string]("gamma", "gamma", false, false)

	or := NewOrParser[string](aParser, bParser, gParser)

	// Build charMap: 'a'->0, 'b'->1, 'g'->2
	var cm [256][]int
	cm['a'] = []int{0}
	cm['b'] = []int{1}
	cm['g'] = []int{2}
	or.SetCharMap(cm)

	_ = identity

	tests := []struct {
		input string
		ok    bool
		match string
	}{
		{"alpha", true, "alpha"},
		{"beta", true, "beta"},
		{"gamma", true, "gamma"},
		{"delta", false, ""},
		{"", false, ""},
	}

	for _, tt := range tests {
		scanner := NewScanner[string](tt.input, nil)
		node, err := ParsePartial(or, scanner)
		if tt.ok {
			if err != nil {
				t.Errorf("input %q: expected match, got error", tt.input)
			} else if node.Payload != tt.match {
				t.Errorf("input %q: expected %q, got %q", tt.input, tt.match, node.Payload)
			}
		} else {
			if err == nil {
				t.Errorf("input %q: expected error, got match %q", tt.input, node.Payload)
			}
		}
	}
}

func TestOrParserCharMapWithWhitespace(t *testing.T) {
	selectP := NewAtomParser[string]("SELECT", "SELECT", true, true)
	insertP := NewAtomParser[string]("INSERT", "INSERT", true, true)
	deleteP := NewAtomParser[string]("DELETE", "DELETE", true, true)

	or := NewOrParser[string](selectP, insertP, deleteP)

	// Both 's'/'S' -> 0, 'i'/'I' -> 1, 'd'/'D' -> 2
	var cm [256][]int
	cm['s'] = []int{0}
	cm['S'] = []int{0}
	cm['i'] = []int{1}
	cm['I'] = []int{1}
	cm['d'] = []int{2}
	cm['D'] = []int{2}
	or.SetCharMap(cm)

	tests := []struct {
		input string
		ok    bool
		match string
	}{
		{"SELECT", true, "SELECT"},
		{"  SELECT", true, "SELECT"},
		{"INSERT", true, "INSERT"},
		{"DELETE", true, "DELETE"},
		{"  delete", true, "DELETE"},
		{"UPDATE", false, ""},
	}

	for _, tt := range tests {
		scanner := NewScanner[string](tt.input, SkipWhitespaceRegex)
		node, err := ParsePartial(or, scanner)
		if tt.ok {
			if err != nil {
				t.Errorf("input %q: expected match, got error", tt.input)
			} else if node.Payload != tt.match {
				t.Errorf("input %q: expected %q, got %q", tt.input, tt.match, node.Payload)
			}
		} else {
			if err == nil {
				t.Errorf("input %q: expected error, got match %q", tt.input, node.Payload)
			}
		}
	}
}
