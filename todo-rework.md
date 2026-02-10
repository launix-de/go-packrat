# go-packrat rework: remaining items (v3)

Context: memcp uses go-packrat for SQL parsing. v2.1.16-v2.1.18 reduced allocations by 99.6% through internal optimizations (no API changes). The remaining items require breaking API changes and are planned for v3.

## Done in v2

- [x] Item 1: AtomParser string compare instead of regex (v2.1.16)
- [x] Item 2: MatchRegexp: FindStringIndex instead of FindStringSubmatch (v2.1.17)
- [x] Item 3: Scanner.breaks []bool instead of map[int]bool (v2.1.16)
- [x] Item 4: Whitespace skip fast-path (v2.1.16)
- [x] Item 5: Pool Lr objects via sync.Pool (v2.1.16)
- [x] Item 6: Flat memoization structure (v2.1.16)
- [x] Item 7: CharMap dispatch for OrParser — SetCharMap() (v2.1.18)
- [x] Item 8: Memoization bypass for Kleene/Many — NoMemo flag (v2.1.18)
- [x] Item 9: MemoEntry slab allocator + linked-list memo (v2.1.17)
- [x] Item 10b: Combinator buffer reuse with depth counter (v2.1.17)
- [x] Item 11: Scanner.Reset() for pooling (v2.1.18)
- [x] Item 12: RegexParser fast-path specialization (v2.1.16)
- [x] Heads map → slice (v2.1.18)

## v3: Merge callback with Scanner reference (item 10)

**Files:** and.go, kleene.go, many.go, scanner.go
**Current:** The Merge callback signature is `func(string, ...T) T`. It receives the matched string and sub-results but has no access to per-query state (Scanner, memoization, arena).
**Change:**
- Add `UserData any` field to `Scanner[T]` struct. This is a caller-defined opaque value, accessible from Merge callbacks
- Change callback signature to `func(*Scanner[T], string, ...T) T`
- Update all combinator `Match` methods to pass `s` as first arg

**Why this matters:** memcp needs to pass a per-query arena allocator through to `mergeParserResults`. Without Scanner access in the callback, the only options are goroutine-local globals or closure captures at parser-construction time — both are either racy or impossible (parsers are constructed once, reused across queries). With Scanner access, the callback reads `s.UserData.(*parseArena)` to arena-allocate `[]Scmer` slices and `*parserResult` objects.

**Impact:** Enables the arena allocator pattern that replaces ~16000 individual `[]Scmer` and `*parserResult` allocations with 1-2 bulk allocations per query (amortized to 0 with pooling).

**Breaking change:** All callers must update their callback signature.
