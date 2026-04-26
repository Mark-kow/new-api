## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.
## 2025-02-23 - [Proxy Payload Accumulation Memory Overhead]
**Learning:** Using `strings.Join` combined with string concatenation (e.g., `"[" + strings.Join(items, ",") + "]"`) for constructing large JSON payloads in high-frequency proxy endpoints causes significant intermediate memory allocations and GC pressure. This is a common pattern in the codebase when parsing or modifying stream items.
**Action:** Always calculate total capacity and use `strings.Builder` with `Grow()` when constructing large JSON array strings from slices of items, bypassing `strings.Join`.
