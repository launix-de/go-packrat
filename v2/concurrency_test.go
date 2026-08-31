/*
	(c) 2026 Launix, Inh. Carl-Philip Hänsch

	Dual licensed with custom agreements or GPLv3
*/

package packrat

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestSharedGrammarConcurrentParsing(t *testing.T) {
	word := NewRegexParser(func(value string) string { return value }, `[a-z]+`, false, false)
	comma := NewAtomParser(",", ",", false, false)
	words := NewManyParser(func(_ string, values ...string) string {
		return strings.Join(values, ",")
	}, word, comma)
	wrapped := NewAndParser(func(_ string, values ...string) string { return values[1] },
		NewAtomParser("[", "[", false, false),
		words,
		NewAtomParser("]", "]", false, false))
	grammar := NewOrParser[string](wrapped, word)

	const workers = 32
	const iterations = 200
	var wg sync.WaitGroup
	errors := make(chan error, workers)
	for worker := 0; worker < workers; worker++ {
		worker := worker
		wg.Add(1)
		go func() {
			defer wg.Done()
			for iteration := 0; iteration < iterations; iteration++ {
				value := "alpha"
				if worker%2 != 0 {
					value = "beta"
				}
				input := "[worker," + value + "]"
				node, parseErr := Parse(grammar, NewScanner[string](input, nil))
				if parseErr != nil {
					errors <- parseErr
					return
				}
				if node.Payload != "worker,"+value {
					errors <- fmt.Errorf("parse %q returned %q", input, node.Payload)
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		t.Fatal(err)
	}
}

func TestTerminalFastPathPreservesFailureDetails(t *testing.T) {
	want := NewAtomParser("expected", "expected", false, false)
	_, parseErr := Parse[string](want, NewScanner[string]("actual", nil))
	if parseErr == nil {
		t.Fatal("expected parse error")
	}
	if parseErr.Position != 0 {
		t.Fatalf("failure position = %d, want 0", parseErr.Position)
	}
	if len(parseErr.FailedParsers) != 1 || parseErr.FailedParsers[0] != want {
		t.Fatalf("failed parsers = %#v, want terminal parser", parseErr.FailedParsers)
	}
	if !strings.Contains(parseErr.Error(), "expected") {
		t.Fatalf("error does not contain expected token: %s", parseErr)
	}
}
