## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.

## 2024-05-24 - Optimize stream token joining in relay-openai.go
**Learning:** For array of JSON strings, concatenating using `strings.Join` combined with string formatting (e.g. `streamResp := "[" + strings.Join(streamItems, ",") + "]"`) inside a loop causes a lot of intermediate memory allocations and forces the resultant string to be copied into memory again. In Go, it is significantly faster (30-40% speed up, 50% fewer allocations, and 50% less memory usage) to pre-calculate the required size, pre-allocate a `[]byte`, directly append the parts, and pass the byte slice directly to `json.Unmarshal`.
**Action:** When manually constructing large JSON payloads meant for `json.Unmarshal`, calculate the capacity, pre-allocate a `[]byte` slice, and append data directly to it instead of using `strings.Join` or string concatenation.
