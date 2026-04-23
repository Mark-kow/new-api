## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.
## 2024-04-23 - JSON concatenation optimization
**Learning:** In Go, dynamically building JSON strings by repeatedly concatenating arrays with `strings.Join` is inefficient. Pre-allocating a `strings.Builder` when appending to an array format avoids huge intermediate allocations.
**Action:** Use `strings.Builder` to construct dynamic JSON payload strings rather than array concatenation.
