## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.
## 2024-11-20 - Avoid `fmt.Sprint` for slice stringification
**Learning:** In hot API gateway paths (like `controller/relay.go`), using convoluted string manipulations on slices (e.g., using `fmt.Sprint` to stringify a slice, then parsing with `strings.Fields` and `strings.Trim`) is an anti-pattern that creates unnecessary allocations and CPU overhead.
**Action:** Use `strings.Join(slice, "->")` directly to join slice elements, which avoids intermediate string parsing and arrays, drastically reducing allocations and improving performance.
