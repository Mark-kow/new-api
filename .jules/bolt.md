## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.

## 2024-05-24 - Pre-allocate strings.Builder for parsing and masking urls
**Learning:** Naively allocating a slice to hold masked items to be joined by `strings.Join` and concatenated to a `result` string causes unnecesary intermediate allocations.
**Action:** Replace `strings.Join(...)` with a pre-allocated `strings.Builder`. Iterating over items, writing them with `sb.WriteString(...)` directly eliminates the need to hold intermediate masked slices and minimizes heap allocations.
