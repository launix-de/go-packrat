package packrat

type KleeneParser struct {
	subParser, sepParser Parser
}

func NewKleeneParser(subparser Parser, sepparser Parser) *KleeneParser {
	return &KleeneParser{subParser: subparser, sepParser: sepparser}
}

func (p *KleeneParser) Set(embedded Parser, separator Parser) {
	p.subParser = embedded
	p.sepParser = separator
}

// Match matches the embedded parser or the empty string.
func (p *KleeneParser) Match(s *Scanner) (*Scanner, Node) {
	cached, wasCached := s.memoization[s.position][p]
	if wasCached {
		nss, node := cached.Scanner, cached.Node
		return nss, node
	}

	startPosition := s.position
	ns := s.Copy()
	var nodes []Node

	i := 0
	for {
		matchedsep := false
		var sepnode Node
		nss := ns

		if i > 0 && p.sepParser != nil {
			if nss.position >= len(s.memoization) {
				break
			}
			cached, wasCached := s.memoization[nss.position][p.sepParser]

			if wasCached {
				nss, sepnode = cached.Scanner, cached.Node
			} else {
				nss, sepnode = p.sepParser.Match(nss)
				s.memoization[ns.position][p.sepParser] = scannerNode{Scanner: nss, Node: sepnode}

				if nss == nil {
					break
				}
			}

			matchedsep = true
		}
		i++

		nss2 := nss
		if nss2.position >= len(s.memoization) {
			break
		}
		cached, wasCached := s.memoization[nss2.position][p.subParser]

		var (
			node Node
		)
		if wasCached {
			nss2, node = cached.Scanner, cached.Node
		} else {
			nss2, node = p.subParser.Match(nss)
			s.memoization[nss.position][p.subParser] = scannerNode{Scanner: nss2, Node: node}

			if nss2 == nil {
				break
			}
		}

		if matchedsep {
			nodes = append(nodes, sepnode)
		}

		nodes = append(nodes, node)
		ns = nss2
	}

	endPosition := ns.position
	matched := ns.input[startPosition:endPosition]
	r := scannerNode{Scanner: ns, Node: Node{Matched: matched, Parser: p, Children: nodes}}

	s.memoization[s.position][p] = r
	return r.Scanner, r.Node
}
