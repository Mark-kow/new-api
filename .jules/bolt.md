## 2025-03-01 - Avoid strings.Join for large dynamic JSON streams
**Learning:** Constructing JSON byte slices using `strings.Join` combined with string concatenation before parsing with `json.Unmarshal` leads to multiple intermediate memory allocations. `json.Unmarshal` naturally processes `[]byte`.
**Action:** When manually parsing streamed arrays of json tokens, calculate the exact required capacity first, then construct the `[]byte` via appending directly. This improves throughput significantly on hot paths.
