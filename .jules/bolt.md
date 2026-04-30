## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.
## 2025-05-18 - String Concatenation in Loops Causes Bottleneck
**Learning:** Using `text += string` within a loop (such as iterating over `[]string` or `[]interface{}`) causes O(n^2) allocations since strings in Go are immutable. Every concatenation allocates a new string. This is a common bottleneck when processing potentially large unmarshaled JSON arrays like `[]string` and `[]interface{}` often used in text blocks.
**Action:** Always prefer `strings.Join` for slices of strings (`[]string`) and `strings.Builder` for dynamic text construction inside a loop to minimize memory allocations.
