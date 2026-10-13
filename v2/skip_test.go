/*
 (c) 2026 Launix, Inh. Carl-Philip Hänsch
 Dual licensed with custom aggreements or GPLv3
*/

package packrat

import (
	"regexp"
	"testing"
)

func TestSkipAtNonWhitespace(t *testing.T) {
	s := NewScanner[int]("SELECT", SkipWhitespaceRegex)
	posBefore := s.position
	s.Skip()
	if s.position != posBefore {
		t.Error("Skip at non-whitespace position should be a no-op")
	}
}

func TestSkipWithLeadingWhitespace(t *testing.T) {
	s := NewScanner[int]("  hello", SkipWhitespaceRegex)
	s.Skip()
	if s.position != 2 {
		t.Errorf("Skip should advance past whitespace, got position %d", s.position)
	}
}

func TestSkipWithComment(t *testing.T) {
	s := NewScanner[int]("/* comment */ hello", SkipWhitespaceAndCommentsRegex)
	s.Skip()
	if s.position != 14 {
		t.Errorf("Skip should advance past comment and following space, got position %d", s.position)
	}
}

func TestSkipAtEndOfInput(t *testing.T) {
	s := NewScanner[int]("x", SkipWhitespaceRegex)
	s.move(1) // move to end
	posBefore := s.position
	s.Skip()
	if s.position != posBefore {
		t.Error("Skip at end of input should be a no-op")
	}
}

func TestSkipNilRegex(t *testing.T) {
	s := NewScanner[int]("  hello", nil)
	posBefore := s.position
	s.Skip()
	if s.position != posBefore {
		t.Error("Skip with nil regex should be a no-op")
	}
}

func TestSkipCustomPrefixes(t *testing.T) {
	cases := []struct {
		name, input, pattern, remaining string
	}{
		{"leading_line", "-- leading\nword", `^(?:--[^\r\n]*(?:\r\n|\r|\n|$)|[\r\n\t ]+)+`, "word"},
		{"hash", "# leading\nword", `^(?:#[^\n]*\n|\s+)+`, "word"},
		{"unicode", "※※word", `^※+`, "word"},
		{"letter", "IGNORE word", `^IGNORE +`, "word"},
		{"mixed", " \t# first\n-- second\n※word", `^(?:\s+|#[^\n]*\n|--[^\n]*\n|※)+`, "word"},
		{"line_at_eof", "-- final", `^--[^\n]*(?:\n|$)`, ""},
		{"mismatch", "word", `^#+`, "word"},
		{"later_match", "/word #later", `#+`, "/word #later"},
		{"empty_match", "word", `^`, "word"},
		{"unanchored_prefix", "##word", `#+`, "word"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			scanner := NewScanner[int](tc.input, regexp.MustCompile(tc.pattern))
			scanner.Skip()
			if scanner.remainingInput != tc.remaining || scanner.position != len(tc.input)-len(tc.remaining) {
				t.Fatalf("Skip() position=%d remaining=%q; want position=%d remaining=%q", scanner.position, scanner.remainingInput, len(tc.input)-len(tc.remaining), tc.remaining)
			}
		})
	}
}

func TestSkipCustomLeadingAndInterstatementComments(t *testing.T) {
	skipper := regexp.MustCompile(`^(?:--[^\r\n]*(?:\r\n|\r|\n|$)|[\r\n\t ]+)+`)
	grammar := NewAndParser(func(_ string, values ...int) int { return values[0] + values[1] },
		NewAtomParser[int](1, "hello", false, true), NewAtomParser[int](2, "world", false, true))
	node, err := Parse(grammar, NewScanner[int]("-- leading\nhello-- between\nworld-- final", skipper))
	if err != nil || node.Payload != 3 {
		t.Fatalf("comments between terminals: node=%v err=%v", node, err)
	}
}

func TestSkipResetAndCopyCustom(t *testing.T) {
	scanner := NewScanner[int](" word", SkipWhitespaceRegex)
	scanner.Skip()
	custom := regexp.MustCompile(`^#+`)
	scanner.Reset("##word", custom)
	clone := scanner.Copy()
	scanner.Skip()
	clone.Skip()
	if scanner.position != 2 || clone.position != 2 {
		t.Fatalf("custom Reset/Copy: positions %d/%d, want 2/2", scanner.position, clone.position)
	}
	scanner.Reset("  word", nil)
	scanner.Skip()
	if scanner.position != 0 {
		t.Fatalf("nil Reset: position=%d", scanner.position)
	}
	scanner.Reset("/* note */ word", SkipWhitespaceAndCommentsRegex)
	scanner.Skip()
	if scanner.remainingInput != "word" {
		t.Fatalf("default Reset: remaining=%q", scanner.remainingInput)
	}
	scanner.Reset("", custom)
	scanner.Skip()
	if scanner.position != 0 {
		t.Fatalf("empty Reset: position=%d", scanner.position)
	}
}

// Exported defaults may be replaced by a caller; only the original built-in
// regexp objects have a known prefix alphabet.
func TestSkipReassignedDefaultUsesCustomPrefix(t *testing.T) {
	original := SkipWhitespaceRegex
	SkipWhitespaceRegex = regexp.MustCompile(`^#+`)
	defer func() { SkipWhitespaceRegex = original }()
	scanner := NewScanner[int]("##word", SkipWhitespaceRegex)
	scanner.Skip()
	if scanner.position != 2 {
		t.Fatalf("reassigned default: position=%d, want 2", scanner.position)
	}
}

func BenchmarkSkipPrefix(b *testing.B) {
	cases := []struct {
		name, input string
		skipper     *regexp.Regexp
	}{
		{"DefaultMismatch", "word", SkipWhitespaceRegex},
		{"DefaultWhitespace", " \t word", SkipWhitespaceRegex},
		{"DefaultComment", "/* comment */ word", SkipWhitespaceAndCommentsRegex},
		{"CustomMismatch", "word", regexp.MustCompile(`^(?:--[^\n]*(?:\n|$)|\s+)+`)},
		{"CustomLine", "-- comment\nword", regexp.MustCompile(`^(?:--[^\n]*(?:\n|$)|\s+)+`)},
		{"CustomHash", "# comment\nword", regexp.MustCompile(`^#[^\n]*\n`)},
		{"CustomUnicode", "※※word", regexp.MustCompile(`^※+`)},
		{"Nil", "word", nil},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			scanner := NewScanner[int](tc.input, tc.skipper)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				scanner.setPosition(0)
				scanner.Skip()
			}
		})
	}
}
