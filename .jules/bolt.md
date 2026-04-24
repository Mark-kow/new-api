## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.

## 2024-05-24 - Optimize JSON array construction for Unmarshal
**Learning:** In Go, constructing large JSON array strings by concatenating parts via `strings.Join(items, ",")` and string addition (`"[" + joined_string + "]"`) to feed into `json.Unmarshal` is highly inefficient. It allocates a very large string unnecessarily, which is then often casted to `[]byte` for `json.Unmarshal`, resulting in duplicate massive allocations.
**Action:** When building JSON inputs manually for `json.Unmarshal` (e.g., combining JSON-formatted streaming items), calculate the expected length and pre-allocate a `[]byte` slice (`make([]byte, 0, length)`). Manually copy the fragments (`append(buf, item...)`) into the byte buffer. This avoids the intermediary string allocation entirely, significantly saving memory and parse time.
