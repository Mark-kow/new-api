## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.

## 2024-06-25 - Optimize processTokens in relay/channel/openai/helper.go
**Learning:** Constructing a JSON array string dynamically from a slice of strings using `"[" + strings.Join(items, ",") + "]"` results in unnecessary intermediate string allocations. A pre-allocated `strings.Builder` can avoid these allocations completely.
**Action:** Always prefer `strings.Builder` with `Grow` pre-allocation over string concatenation and `strings.Join` when building complex or JSON-like strings dynamically from slices, particularly in hot paths like streaming proxy handlers.
