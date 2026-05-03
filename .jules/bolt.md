## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.
## 2024-05-03 - Optimize JSON streaming builder in openai relay handler
**Learning:** `json.Unmarshal` natively accepts `[]byte`, but when processing concatenated stream chunk items, the previous implementation used `strings.Join` and string concatenations (`"[" + strings.Join(...) + "]"`) before converting to `[]byte` with `common.StringToByteSlice`. This created multiple large hidden allocations proportional to the request body size.
**Action:** When manually constructing large JSON array payloads out of strings meant for `json.Unmarshal`, calculate the capacity, pre-allocate a `[]byte` slice, and append data directly to it instead of using `strings.Join` or string concatenation.
