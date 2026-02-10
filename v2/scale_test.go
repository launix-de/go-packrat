package packrat

import (
	"fmt"
	"testing"
)

func benchInsertN(b *testing.B, n int) {
	identity := func(s string) string { return s }
	cb := func(match string, a ...string) string { return match }

	lparen := NewAtomParser[string]("(", "(", false, false)
	rparen := NewAtomParser[string](")", ")", false, false)
	comma := NewAtomParser[string](",", ",", false, false)
	intP := NewRegexParser(identity, `-?[0-9]+`, false, false)
	strBody := NewRegexParser(identity, `(\\.|[^\\'])*`, false, false)
	quote := NewAtomParser[string]("'", "'", false, false)

	strP := NewAndParser(cb, quote, strBody, quote)
	value := NewOrParser[string](intP, strP)
	valueList := NewManyParser(cb, value, comma)
	tuple := NewAndParser(cb, lparen, valueList, rparen)
	tuples := NewManyParser(cb, tuple, comma)

	input := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			input += ","
		}
		input += "(1,'hello')"
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner := NewScanner[string](input, nil)
		_, err := ParsePartial(tuples, scanner)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkScale10(b *testing.B)    { benchInsertN(b, 10) }
func BenchmarkScale50(b *testing.B)    { benchInsertN(b, 50) }
func BenchmarkScale200(b *testing.B)   { benchInsertN(b, 200) }
func BenchmarkScale1000(b *testing.B)  { benchInsertN(b, 1000) }
func BenchmarkScale2000(b *testing.B)  { benchInsertN(b, 2000) }

func TestAllocScaling(t *testing.T) {
	identity := func(s string) string { return s }
	cb := func(match string, a ...string) string { return match }

	lparen := NewAtomParser[string]("(", "(", false, false)
	rparen := NewAtomParser[string](")", ")", false, false)
	comma := NewAtomParser[string](",", ",", false, false)
	intP := NewRegexParser(identity, `-?[0-9]+`, false, false)
	strBody := NewRegexParser(identity, `(\\.|[^\\'])*`, false, false)
	quote := NewAtomParser[string]("'", "'", false, false)

	strP := NewAndParser(cb, quote, strBody, quote)
	value := NewOrParser[string](intP, strP)
	valueList := NewManyParser(cb, value, comma)
	tuple := NewAndParser(cb, lparen, valueList, rparen)
	tuples := NewManyParser(cb, tuple, comma)

	fmt.Println("--- NewScanner each time ---")
	for _, n := range []int{1, 10, 50, 200, 1000, 2000} {
		input := ""
		for i := 0; i < n; i++ {
			if i > 0 {
				input += ","
			}
			input += "(1,'hello')"
		}
		result := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				scanner := NewScanner[string](input, nil)
				ParsePartial(tuples, scanner)
			}
		})
		fmt.Printf("tuples=%4d  len=%6d  allocs=%4d  bytes=%8d  ns/op=%8d\n",
			n, len(input), result.AllocsPerOp(), result.AllocedBytesPerOp(), result.NsPerOp())
	}

	fmt.Println("--- Scanner.Reset (reuse) ---")
	scanner := NewScanner[string]("", nil)
	for _, n := range []int{1, 10, 50, 200, 1000, 2000} {
		input := ""
		for i := 0; i < n; i++ {
			if i > 0 {
				input += ","
			}
			input += "(1,'hello')"
		}
		result := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				scanner.Reset(input, nil)
				ParsePartial(tuples, scanner)
			}
		})
		fmt.Printf("tuples=%4d  len=%6d  allocs=%4d  bytes=%8d  ns/op=%8d\n",
			n, len(input), result.AllocsPerOp(), result.AllocedBytesPerOp(), result.NsPerOp())
	}

	fmt.Println("--- Scanner.Reset + NoMemo ---")
	tuples.NoMemo = true
	valueList.NoMemo = true
	for _, n := range []int{1, 10, 50, 200, 1000, 2000} {
		input := ""
		for i := 0; i < n; i++ {
			if i > 0 {
				input += ","
			}
			input += "(1,'hello')"
		}
		result := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				scanner.Reset(input, nil)
				ParsePartial(tuples, scanner)
			}
		})
		fmt.Printf("tuples=%4d  len=%6d  allocs=%4d  bytes=%8d  ns/op=%8d\n",
			n, len(input), result.AllocsPerOp(), result.AllocedBytesPerOp(), result.NsPerOp())
	}
	tuples.NoMemo = false
	valueList.NoMemo = false
}
