## 2024-05-20 - Silent Failure Anti-Pattern in DeleteUser
**Vulnerability:** In `controller/user.go` `DeleteUser`, when `model.HardDeleteUserById(id)` fails, the response is mistakenly returned with `"success": true` and empty message. Also, when it succeeds, it misses returning success true since `err == nil` skips the block and the function returns empty 200 without JSON content, so `success` is not returned when it succeeds!
**Learning:** `err != nil` returning `success: true` is a silent failure anti-pattern mentioned in memory. Furthermore, the successful path implicitly ends without rendering JSON, causing an empty response.
**Prevention:** Proper error handling must use `common.ApiError` and return. The successful path must output a valid `success: true` JSON payload.
