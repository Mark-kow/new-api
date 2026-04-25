## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.

## 2026-04-25 - Avoid `strings.Join` for intermediate JSON string construction
**Learning:** Constructing large JSON strings for `json.Unmarshal` using string concatenation like `"[" + strings.Join(items, ",") + "]"` results in excessive memory allocation due to string immutability and subsequent conversion to `[]byte`.
**Action:** When manually formatting arrays or large payloads intended for JSON deserialization, calculate the final capacity, pre-allocate a `[]byte` slice, and append directly using `append(slice, ...)`. This avoids intermediate string allocations and the overhead of `StringToByteSlice` type conversions.
