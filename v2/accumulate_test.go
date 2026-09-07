/*
	(c) 2026 Launix, Inh. Carl-Philip Hänsch

	Dual licensed with custom aggreements or GPLv3
*/

package packrat

import (
	"strconv"
	"testing"
)

// The accumulation form of Kleene / Many folds every matched item into an
// accumulator (Init / Step / Finish) instead of collecting them for callback.

func numberParser() Parser[int] {
	return NewRegexParser(func(s string) int { n, _ := strconv.Atoi(s); return n }, "[0-9]+", false, true)
}

// TestSumExpressionNoSlice parses "1+2+3" straight to 6 without ever
// materialising the [1 2 3] operand slice: Step adds each operand into a
// running int as it is matched.
func TestSumExpressionNoSlice(t *testing.T) {
	plus := NewAtomParser(0, "+", false, true)

	expr := NewManyParser[int](nil, numberParser(), plus)
	expr.SetAccumulator(
		func() int { return 0 },
		func(sum, operand int) int { return sum + operand },
		nil,
	)

	n, err := Parse[int](expr, NewScanner[int]("1+2+3", SkipWhitespaceRegex))
	if err != nil {
		t.Fatal(err)
	}
	if n.Payload != 6 {
		t.Fatalf("1+2+3 folded to %d, want 6", n.Payload)
	}

	// a single operand: no separator, one Step, still the operand itself
	one, err := Parse[int](expr, NewScanner[int]("42", SkipWhitespaceRegex))
	if err != nil {
		t.Fatal(err)
	}
	if one.Payload != 42 {
		t.Fatalf("single operand folded to %d, want 42", one.Payload)
	}
}

func TestKleeneAccumulate(t *testing.T) {
	sep := NewAtomParser(0, ",", false, true)

	sum := NewKleeneParser[int](nil, numberParser(), sep)
	sum.SetAccumulator(
		func() int { return 100 },                  // seed
		func(acc, item int) int { return acc + item }, // fold
		func(acc int) int { return acc * 2 },        // finish
	)

	n, err := Parse[int](sum, NewScanner[int]("1,2,3", SkipWhitespaceRegex))
	if err != nil {
		t.Fatal(err)
	}
	if n.Payload != (100+1+2+3)*2 {
		t.Fatalf("Kleene accumulate = %d, want %d", n.Payload, (100+1+2+3)*2)
	}

	// Kleene matches the empty string: Init + Finish still run, no Step.
	empty, err := ParsePartial[int](sum, NewScanner[int]("x", SkipWhitespaceRegex))
	if err != nil {
		t.Fatal(err)
	}
	if empty.Payload != 100*2 {
		t.Fatalf("empty Kleene accumulate = %d, want %d", empty.Payload, 100*2)
	}
}

func TestManyAccumulate(t *testing.T) {
	sep := NewAtomParser(0, ",", false, true)

	count := NewManyParser[int](nil, numberParser(), sep)
	count.SetAccumulator(
		func() int { return 0 },
		func(acc, item int) int { return acc + 1 },
		nil,
	)

	n, err := Parse[int](count, NewScanner[int]("7,7,7,7", SkipWhitespaceRegex))
	if err != nil {
		t.Fatal(err)
	}
	if n.Payload != 4 {
		t.Fatalf("Many accumulate count = %d, want 4", n.Payload)
	}

	// Many still fails when nothing matched.
	if _, err := Parse[int](count, NewScanner[int]("nope", SkipWhitespaceRegex)); err == nil {
		t.Fatal("Many accumulate should fail on zero matches")
	}
}

// An accumulating repeat nested inside a backtracking alternative must re-init
// on every entry - the accumulator is never shared across attempts.
func TestAccumulateBacktrackReinit(t *testing.T) {
	sep := NewAtomParser(0, ",", false, true)
	nums := NewManyParser[int](nil, numberParser(), sep)
	nums.SetAccumulator(
		func() int { return 0 },
		func(acc, item int) int { return acc + item },
		nil,
	)
	// (nums "!") | (nums "?")
	bang := NewAndParser(func(s string, a ...int) int { return a[0] }, nums, NewAtomParser(0, "!", false, true))
	quest := NewAndParser(func(s string, a ...int) int { return a[0] }, nums, NewAtomParser(0, "?", false, true))
	alt := NewOrParser[int](bang, quest)

	n, err := Parse[int](alt, NewScanner[int]("1,2,3?", SkipWhitespaceRegex))
	if err != nil {
		t.Fatal(err)
	}
	if n.Payload != 6 {
		t.Fatalf("re-init accumulate = %d, want 6 (not 12)", n.Payload)
	}
}
