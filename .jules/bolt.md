## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.
## 2026-04-28 - Optimize processTokens in relay-openai helper.go
**Learning:** In Go, when manually constructing large JSON payloads meant for `json.Unmarshal`, calculate the capacity, pre-allocate a `[]byte` slice, and append data directly to it instead of using `strings.Join` or string concatenation. This avoids unnecessary intermediate string allocations and subsequent conversion overhead.
**Action:** Replace `strings.Join` combined with string concatenation for constructing dynamic JSON arrays with manual byte slice construction and pre-allocation based on expected lengths.
