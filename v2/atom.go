/*
	(c) 2019 Launix, Inh. Carl-Philip Hänsch
	Author: Tim Kluge

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import (
	"regexp"
)

type AtomParser[T any] struct {
	value T
	r      *regexp.Regexp
	skipWs bool
	atom   string
}

func NewAtomParser[T any](value T, str string, caseInsensitive bool, skipWs bool) *AtomParser[T] {
	prefix := ""
	if caseInsensitive {
		prefix += "(?i)"
	}
	prefix += "^"
	r := regexp.MustCompile(prefix + regexp.QuoteMeta(str))
	p := &AtomParser[T]{value: value, skipWs: skipWs, r: r, atom: str}
	return p
}

// Match matches only the given string. If skipWs is set to true, leading whitespace according to the scanner's skip regexp is skipped, but not matched by the parser.
func (p *AtomParser[T]) Match(s *Scanner[T]) (Node[T], bool) {
	startPosition := s.position

	if p.skipWs {
		s.Skip()

		if !s.isAtBreak() {
			s.setPosition(startPosition)
			return Node[T]{}, false
		}
	}

	matched := s.MatchRegexp(p.r)
	if matched == nil {
		s.setPosition(startPosition)
		return Node[T]{}, false
	}

	if p.skipWs {
		if !s.isAtBreak() {
			s.setPosition(startPosition)
			return Node[T]{}, false
		}
	}

	return Node[T]{Payload: p.value}, true
}
