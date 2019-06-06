/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import "regexp"

type RegexParser struct {
	regex  *regexp.Regexp
	skipWs bool
}

func NewRegexParser(rs string, caseInsensitive bool, skipWs bool) *RegexParser {
	prefix := ""
	if caseInsensitive {
		prefix += "(?i)"
	}
	prefix += "^"
	r := regexp.MustCompile(prefix + rs)
	return &RegexParser{regex: r, skipWs: skipWs}
}

// Regex matches only the given regexp. If skipWs is set to true, leading whitespace according to the scanner's skip regexp is skipped, but not matched by the parser.
// Regex panics if rs is not a valid regex string.
func (p *RegexParser) Match(s *Scanner) (*Scanner, Node) {
	if p.skipWs {
		s.Skip()
		if !s.isAtBreak() {
			return nil, Node{}
		}
	}


	matched := s.MatchRegexp(p.regex)
	if matched == nil {
		return nil, Node{}
	}

	if p.skipWs {
		if !s.isAtBreak() {
			return nil, Node{}
		}
	}

	return s, Node{Matched: *matched, Parser: p}
}
