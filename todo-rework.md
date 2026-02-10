# go-packrat rework: performance and allocation reduction

Context: memcp uses go-packrat for SQL parsing. Large INSERT statements (2000+ value tuples) cause ~80000 heap allocations inside the packrat lib alone. This rework targets near-zero-allocation parsing for repetitive grammars.

## 2. MatchRegexp: use FindStringIndex instead of FindStringSubmatch

**File:** scanner.go:208-216
**Current:** `FindStringSubmatch(s.remainingInput)` returns `[]string` (heap allocation per call).
**Change:** Use `FindStringIndex(s.remainingInput)` which returns `[]int{start, end}`. Extract the matched text via `s.remainingInput[:end]` (no allocation, Go strings share backing array). Return the match length instead of a `*string`.
**Signature change:** `MatchRegexp(r *regexp.Regexp) (int, bool)` instead of `*string`. Callers that need the matched text use `s.input[startPos:startPos+matchLen]`.

**Impact:** Eliminates string allocation per regex match (~4000 allocs for RegexParser in INSERT). `FindStringIndex` also avoids capturing group processing.

## 7. Character map dispatch for OrParser

**File:** or.go
**Current:** `OrParser.Match` tries each sub-parser sequentially (line 25-31).
**Change:** Add an optional `charMap [256][]int` field to `OrParser`. When populated:
```go
func (p *OrParser[T]) Match(s *Scanner[T]) (Node[T], bool) {
    if p.charMap != nil {
        s.Skip()
        if s.position >= len(s.input) {
            return Node[T]{}, false
        }
        candidates := p.charMap[s.input[s.position]]
        startPosition := s.position
        for _, idx := range candidates {
            node, ok := s.applyRule(p.subParser[idx])
            if ok {
                return Node[T]{Payload: node.Payload}, true
            }
            s.setPosition(startPosition)
        }
        return Node[T]{}, false
    }
    // ... existing sequential fallback ...
}
```
Add `SetCharMap(cm [256][]int)` method. The char map is built externally (by memcp's `OptimizeParser()`) based on analysis of each sub-parser's first-character set.

**Impact:** Reduces N sequential match attempts to 1-2 for keyword-prefixed alternatives. For `sql_statement` with ~20 alternatives, saves ~19 failed atom regex matches per statement parse.

**Build helper:** Consider adding a `FirstCharSet() [256]bool` method to the `Parser[T]` interface (or as a standalone analysis function) that computes which first bytes a parser can match. `AtomParser`: exact chars from atom string. `RegexParser`: analyze regex prefix. `AndParser`: delegate to first sub-parser. `OrParser`: union of all sub-parsers. Unknown: all-true (conservative).

## 8. Memoization bypass for Kleene/Many bodies

**Files:** kleene.go, many.go
**Current:** The loop body calls `s.applyRule(p.subParser)` and `s.applyRule(p.sepParser)` which create memo entries per position. For 2000 iterations this creates ~10000 memo map entries that are never looked up again (the parser never backtracks to position 5000 after reaching position 15000).
**Change:** Add a `noMemo bool` flag to Kleene/ManyParser. When true, call `p.subParser.Match(s)` and `p.sepParser.Match(s)` directly instead of `s.applyRule(...)`. This bypasses the memo table entirely.
- Set `noMemo = true` when the sub-parser is known to be non-left-recursive (which is true for all SQL expression grammars in practice)
- Conservative default: `noMemo = false` (preserves correctness for left-recursive grammars)
- The flag can be set by the caller (memcp's `OptimizeParser()`) or auto-detected by checking if the sub-parser graph contains cycles

**Impact:** Eliminates ~58000 memo-related allocations (Lr, MemoEntry, memo maps) for the 2000-tuple INSERT. This is the single largest allocation reduction.

**Caveat:** Bypassing memoization means repeated failures at the same position will re-execute the parser instead of returning the cached failure. For Kleene/Many bodies this is fine because the loop only continues on success (failures immediately break the loop). The separator parser failure that terminates the loop happens once and is cheap.

## 10. Merge callback: pass Scanner reference

**Files:** and.go, kleene.go, many.go, scanner.go
**Current:** The Merge callback signature is `func(string, ...T) T`. It receives the matched string and sub-results but has no access to per-query state (Scanner, memoization, arena).
**Change:**
- Add `UserData any` field to `Scanner[T]` struct (scanner.go:61-74). This is a caller-defined opaque value, accessible from Merge callbacks
- Change callback signature to `func(*Scanner[T], string, ...T) T`
- Update all combinator `Match` methods to pass `s` as first arg:
  ```go
  // and.go:41 (current)
  return Node[T]{Payload: p.callback(s.input[start:s.position], nodes...)}, true
  // and.go:41 (new)
  return Node[T]{Payload: p.callback(s, s.input[start:s.position], nodes...)}, true
  ```
- Same change in kleene.go:51,53 and many.go (equivalent lines)

**Why this matters:** memcp needs to pass a per-query arena allocator through to `mergeParserResults`. Without Scanner access in the callback, the only options are goroutine-local globals or closure captures at parser-construction time — both are either racy or impossible (parsers are constructed once, reused across queries). With Scanner access, the callback reads `s.UserData.(*parseArena)` to arena-allocate `[]Scmer` slices and `*parserResult` objects.

**Impact:** Enables the arena allocator pattern that replaces ~16000 individual `[]Scmer` and `*parserResult` allocations with 1-2 bulk allocations per query (amortized to 0 with pooling). This is the prerequisite for memcp's allocation-free parsing wrapper.

**Breaking change:** All callers of `NewAndParser`, `NewKleeneParser`, `NewManyParser` must update their callback signature. For memcp this means updating `mergeParserResults` and `mergeParserResultsNil` in packrat.go. For other users of go-packrat: add `_ *Scanner[T]` as first parameter to their callbacks.

## 11. Scanner pooling with Reset method

**File:** scanner.go
**Current:** `NewScanner()` (scanner.go:169-192) allocates a new Scanner struct with new maps every time. For repeated parsing (e.g., the SQL parser called per query), this creates per-query allocations for the Scanner struct, breaks map/slice, memoization map, and heads map.
**Change:** Add a `Reset(input string, skipper *regexp.Regexp)` method:
```go
func (s *Scanner[T]) Reset(input string, skipper *regexp.Regexp) {
    s.input = input
    s.position = 0
    s.remainingInput = input
    s.skipRegex = skipper
    s.invocationStack = nil

    // Reuse breaks slice if large enough (requires item 3: []bool)
    needed := len(input) + 1
    if cap(s.breaks) >= needed {
        s.breaks = s.breaks[:needed]
        // zero out
        for i := range s.breaks { s.breaks[i] = false }
    } else {
        s.breaks = make([]bool, needed)
    }
    // rebuild breaks (same loop as NewScanner)
    previousWord := false
    for pos, r := range input {
        currentWord := unicode.In(r, unicode.N, unicode.L, unicode.Pc)
        if !currentWord || !previousWord {
            s.breaks[pos] = true
        }
        previousWord = currentWord
    }
    s.breaks[len(input)] = true

    // Reuse maps: clear() keeps allocated buckets
    clear(s.memoization)
    clear(s.heads)
}
```

The caller (memcp's `ScmParser.Execute()`) pools Scanners via `sync.Pool`:
```go
var scannerPool = sync.Pool{New: func() any {
    return packrat.NewScanner[*parserResult]("", nil)
}}

func (b *ScmParser) Execute(str string, en *Env) Scmer {
    scanner := scannerPool.Get().(*packrat.Scanner[*parserResult])
    scanner.Reset(str, skipper)
    defer scannerPool.Put(scanner)
    // ...
}
```

**Impact:** Eliminates per-query Scanner struct allocation, breaks slice allocation (amortized for similar-length inputs), and map bucket re-allocation. Combined with item 10 (UserData for arena), the Scanner+arena can be pooled as a unit.

**Dependency:** Item 3 (breaks as `[]bool`) is done. Reset can reuse the slice.

## Priority order

| # | Task | Effort | Impact | Dependencies |
|---|------|--------|--------|-------------|
| 1 | MatchRegexp: FindStringIndex (item 2) | Low | Medium | None |
| 2 | Memo bypass in Kleene/Many (item 8) | Medium | Very High | None |
| 3 | Merge callback with Scanner ref (item 10) | Low | High (enabler) | None |
| 4 | Scanner pooling with Reset (item 11) | Low | Medium | None |
| 5 | CharMap dispatch for OrParser (item 7) | Medium | Medium | None |
