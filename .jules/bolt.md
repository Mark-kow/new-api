## 2024-05-24 - Optimize unescapeString in relay-gemini.go
**Learning:** In Go, string concatenation by repeatedly appending to a `[]rune` slice can be a bottleneck. Using `strings.Builder` with `builder.Grow()` pre-allocation is significantly faster and uses less memory.
**Action:** When manipulating strings in performance-critical paths (like parsing JSON or handling escape sequences), favor `strings.Builder` and add fast-paths to skip processing when no modification is needed.

## 2024-05-24 - Avoid abstraction leaks for micro-optimizations
**Learning:** Optimizing "cold paths" (e.g. admin-facing functions) by bypassing domain abstractions (like `GetEnabledModels()`) with raw SQL queries breaks encapsulation and introduces potential logic bugs for negligible real-world gain.
**Action:** Always consider the business context and hit frequency of the path being optimized. Never sacrifice maintainability or bypass core abstractions unless the path is confirmed as a critical bottleneck.

## 2024-05-24 - Simplify string serialization
**Learning:** In hot API gateway paths (like the relay controller), complex string manipulations such as `fmt.Sprintf("...", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(slice)), "->"), "[]"))` create multiple unnecessary slice and string allocations.
**Action:** Replace convoluted string serializations with direct and simple standard library calls, like `strings.Join(slice, "->")`, which reduces memory overhead and improves processing speed.
